# Design: Add Feed System

## Overview

The current codebase already supports a homepage content stream, but the behavior is implemented as a content-list flow:

```text
frontend
  -> /content/post/list
      -> content-api
          -> content-rpc
              -> MySQL post table
              -> optional cache for latest feed
frontend
  -> /user/batch
  -> /interaction/post/star/batch
  -> /interaction/post/collection/batch
  -> /content/post/batch
```

This change does not implement a new runtime architecture. It defines the target feed capability and describes how the current implementation should evolve toward it.

## Current State

The current homepage behavior is split across multiple services:

- `content` stores and lists posts
- `relation` stores following relationships
- `interaction` stores viewer actions and counters
- `user` provides author enrichment
- `search` and Canal/Elasticsearch provide searchable read models

The current list behavior already supports:

- `latest`
- `following`

The current RPC implementation for `following` is a read-time join between `post` and `following`, while `latest` is a mostly time-ordered content list with cache support.

This is a valid early-stage design, but it is still a content-list implementation, not a defined feed system.

## Capability Model

The feed capability should answer a separate question from content storage.

- Content asks: what content exists?
- Feed asks: what content should this viewer see, in what order, and with what card data?

This change defines feed as a separate capability even if the first implementation still lives partly inside the `content` service.

## Feed Types

### 1. Latest feed

Purpose:

- primary default feed for anonymous users
- fallback feed for logged-in users
- newest public content discovery

Initial behavior:

- public posts only
- sorted by operator priority first, then time
- supports standard pagination

### 2. Following feed

Purpose:

- primary personalized feed for logged-in users
- shows recent content from followed authors

Initial behavior:

- requires authentication
- includes content from followed authors that the viewer is allowed to see
- sorted by operator priority first, then time

### 3. Future feed types

The contract should be extensible to later support:

- `recommend`
- `user_posts`
- `topic_posts`

These future types are not required for V1 delivery but should not require redesign of the response model.

## Feed Item Model

The feed list should return a stable feed-card shape rather than requiring the frontend to assemble multiple partial models.

```text
FeedItem
  = content snapshot
  + author snapshot
  + aggregate interaction counters
  + current viewer state
  + feed metadata
```

### Required fields

Content snapshot:

- `postId`
- `authorId`
- `title`
- `summary`
- `cover`
- `tags`
- `visibility`
- `isTop`
- `isEssence`
- `isLock`
- `createdAt`
- `updatedAt`
- `latestRepliedAt`

Aggregate counters:

- `commentCount`
- `upvoteCount`
- `collectionCount`
- `shareCount`

Author snapshot:

- `authorName`
- `authorAvatar`

Viewer state:

- `isStarred`
- `isCollected`
- `isFollowingAuthor`

Feed metadata:

- `feedType`
- `source`
- `score` or equivalent ranking marker for future ranked feeds

Not every field must be exposed in the first public API, but this is the target conceptual model for feed cards.

## Responsibility Boundaries

### Content service

Owns:

- post creation
- post update and delete
- post detail
- post content storage
- topic linkage

Does not conceptually own:

- personalized feed ranking
- cross-service feed assembly as a long-term responsibility

### Relation service

Owns:

- follow and unfollow
- follow status
- follower/following lists

### Interaction service

Owns:

- star and collection actions
- share counters
- viewer-specific interaction state

### Feed assembly layer

Owns:

- feed-type dispatch
- candidate retrieval
- visibility filtering
- ranking
- viewer-state enrichment
- feed-card response assembly

In the short term, this assembly may still be implemented inside existing `content` flows. In the long term, it should be treated as a distinct module or service boundary.

## Filtering Rules

Every feed type must apply filtering before ranking and response shaping.

### Visibility

Feed delivery must respect post visibility rules already encoded in the data model, including cases such as:

- private
- followers-only
- mutual-follow-only
- public

### Deletion and invalid state

Feed delivery must exclude:

- soft-deleted posts
- posts that should not be visible due to state or moderation constraints

### Viewer context

Filtering behavior may depend on:

- whether the viewer is logged in
- whether the viewer follows the author
- whether the viewer and author are mutual followers

## Ranking Model

V1 ranking should stay intentionally simple.

### V1 ranking

For latest and following feeds:

1. top or operator-priority content first
2. then reverse chronological order

This matches the current product shape and avoids premature ranking complexity.

### Future ranking shape

The design should preserve a future pipeline:

```text
candidate retrieval
  -> filtering
  -> scoring
  -> mixing
  -> pagination
```

This allows later expansion to:

- recommendation scores
- relationship boosts
- topic affinity
- quality or moderation downranking
- commercial or operator insertion

## Pagination Model

The current system uses page-number pagination. That is acceptable for an initial version, but feed design should anticipate cursor-based pagination.

### V1

- allow page and page size
- maintain deterministic sort order

### Future direction

- move toward cursor-based pagination for better feed continuity
- support rank markers or timestamp-based cursors

This should be treated as an evolution path rather than a V1 blocker.

## API Model

The first public feed contract may be introduced either by evolving the current list endpoint or by defining a feed-specific endpoint.

### Request model

The V1 request shape should support:

- `feedType`
- `page`
- `pageSize`

Example conceptual request:

```json
{
  "feedType": "latest",
  "page": 1,
  "pageSize": 10
}
```

For transition planning, the current `/content/post/list` endpoint may continue to carry this contract as long as its semantics are explicitly documented as feed semantics.

### Response model

The V1 response shape should support:

- `total`
- `list`

Each element in `list` should match the feed item model defined earlier, with V1-mandatory fields present and future ranking metadata optional.

Example conceptual response:

```json
{
  "total": 100,
  "list": [
    {
      "postId": 123,
      "authorId": 9,
      "title": "Example post",
      "summary": "Example summary",
      "cover": "",
      "tags": ["golang", "backend"],
      "visibility": 90,
      "isTop": false,
      "isEssence": false,
      "isLock": false,
      "commentCount": 3,
      "upvoteCount": 12,
      "collectionCount": 4,
      "shareCount": 1,
      "authorName": "alice",
      "authorAvatar": "",
      "isStarred": false,
      "isCollected": false,
      "isFollowingAuthor": true,
      "feedType": "following",
      "source": "following",
      "createdAt": 1700000000,
      "updatedAt": 1700000500,
      "latestRepliedAt": 1700000700
    }
  ]
}
```

## Degraded Experience Model

The current frontend already tolerates partial dependency failure. The feed design should preserve that behavior deliberately rather than leaving it incidental.

### Core dependency

The core dependency is feed-item retrieval itself. If this fails, feed rendering fails.

### Non-core enrichment dependencies

The following should be treated as degradable when the architecture still requires multi-source assembly:

- author snapshot lookup
- summary enrichment
- viewer interaction-state enrichment

Expected behavior:

- do not block feed browsing when non-core enrichment fails,
- expose fallback values where possible,
- make degraded rendering an intentional product behavior.

## Architecture Evolution Strategy

### Phase 1: Define a feed contract on top of the current read path

Keep the current read-time MySQL approach, but reshape the contract around feed semantics.

Goals:

- unify feed item structure
- reduce frontend fan-out where practical
- clarify feed types and behavior

### Phase 2: Isolate feed assembly

Move aggregation, enrichment, and ranking responsibilities behind a more explicit feed boundary.

Possible outcomes:

- a dedicated feed API inside the existing service set
- a separate feed module
- eventually a separate feed service

### Phase 3: Introduce hybrid distribution where justified

The current `following` implementation is read fan-out:

```text
viewer opens feed
  -> query posts
  -> join following
  -> sort and page
```

This is appropriate at the current scale, but future growth may require hybrid distribution.

Future options:

- fan-out on read for simplicity
- fan-out on write for faster read performance
- hybrid strategy for large-author and normal-author separation

This change does not choose a final scaling model. It records the likely evolution path and keeps the V1 design compatible with it.

## Risks

### Risk 1: Over-designing V1

Mitigation:

- keep V1 limited to latest and following
- avoid requiring recommendation infrastructure in this change

### Risk 2: Keeping feed logic scattered across services

Mitigation:

- define feed assembly as an explicit responsibility now
- use the contract to drive later consolidation

### Risk 3: Frontend remains dependent on multiple enrichment endpoints

Mitigation:

- treat unified feed-item assembly as an explicit implementation goal
- record it in tasks rather than assuming current composition is good enough

### Risk 4: Visibility logic becomes inconsistent across list surfaces

Mitigation:

- document visibility as a core feed concern
- require shared filtering semantics for feed variants

### Risk 5: Migration from post-list semantics to feed semantics remains unclear

Mitigation:

- define explicit endpoint transition behavior
- document whether `/content/post/list` is preserved, wrapped, or superseded

## Implementation Follow-up

### Backend follow-up

- centralize feed-card assembly so frontend no longer needs multiple enrichment calls for the normal path
- consolidate visibility filtering into a reusable feed-layer rule set
- define whether the first implementation lives in `content` or a dedicated feed module
- preserve compatibility with the current MySQL read path while making ranked-feed evolution possible

### Frontend follow-up

- switch homepage and related list views toward a unified feed item contract
- reduce view-level knowledge of per-service enrichment calls
- preserve current degraded rendering behavior where backend centralization is not yet complete

### Verification follow-up

- anonymous latest feed returns visible public content
- anonymous following feed is rejected
- authenticated following feed returns followed-author content only
- visibility filtering excludes unauthorized content
- degraded enrichment states still allow feed browsing

## Verification Approach

This design work is complete when:

- the proposal defines feed as a distinct capability,
- the feed types and item contract are clear enough to guide API design,
- the evolution path fits the current codebase without forcing premature re-architecture,
- the tasks can be implemented incrementally later.
