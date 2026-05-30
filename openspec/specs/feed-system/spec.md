# Feed System

## Purpose

The feed system provides list-oriented content delivery for homepage and related discovery surfaces. It defines what feed types the product supports, how feed items are shaped, how visibility and ranking rules are applied, and what contract clients use to request feed data.

## Requirements

### Requirement: Support latest and following feeds

The system SHALL support `latest` and `following` as the first production feed types.

#### Scenario: Anonymous client requests latest feed

- Given a client without authentication
- When the client requests the latest feed
- Then the system returns a feed list of visible public content

#### Scenario: Authenticated client requests latest feed

- Given a logged-in client
- When the client requests the latest feed
- Then the system returns a feed list of visible content using the same latest-feed contract

#### Scenario: Authenticated client requests following feed

- Given a logged-in client
- When the client requests the following feed
- Then the system returns recent visible content authored by followed users

#### Scenario: Anonymous client requests following feed

- Given a client without authentication
- When the client requests the following feed
- Then the system rejects the request as unauthorized

### Requirement: Keep the feed contract extensible

The system SHALL preserve a feed-type model that can later support additional feed variants without redesigning the core response shape.

#### Scenario: Future feed types are added

- Given a later product phase
- When the system introduces feed types such as `recommend`, `user_posts`, or `topic_posts`
- Then the request and response model can be extended without redefining the basic feed item shape

### Requirement: Expose a stable feed item model

The system SHALL return feed items as stable card-oriented objects suitable for direct rendering.

#### Scenario: Feed item includes content snapshot

- Given a successful feed-list response
- Then each feed item includes content fields such as `postId`, `authorId`, `title`, `summary`, `cover`, `tags`, `visibility`, `isTop`, `isEssence`, `isLock`, `createdAt`, `updatedAt`, and `latestRepliedAt`

#### Scenario: Feed item includes aggregate counters

- Given a successful feed-list response
- Then each feed item includes aggregate counters such as `commentCount`, `upvoteCount`, `collectionCount`, and `shareCount`

#### Scenario: Feed item includes author snapshot

- Given a successful feed-list response
- Then each feed item includes author display fields such as `authorName` and `authorAvatar`

#### Scenario: Feed item includes viewer state when available

- Given an authenticated viewer
- When the system returns feed items
- Then each feed item may include viewer-state fields such as `isStarred`, `isCollected`, and `isFollowingAuthor`

#### Scenario: Feed item includes feed metadata

- Given a successful feed-list response
- Then each feed item includes feed metadata such as `feedType` and `source`
- And the model reserves a ranking marker such as `score` for future ranked feeds

### Requirement: Distinguish mandatory V1 fields from future ranking fields

The system SHALL guarantee the minimum field set required for V1 rendering and SHALL reserve future ranking fields as optional extensions.

#### Scenario: V1 clients receive mandatory fields

- Given a V1 feed-list response
- Then clients receive the fields needed to render content, author, counters, and current feed context

#### Scenario: Future ranked feeds add optional metadata

- Given a future ranked feed
- When ranking metadata becomes available
- Then fields such as `score` may be added without breaking V1 clients that do not depend on them

### Requirement: Support graceful degraded feed rendering

The system SHALL define degraded behavior when non-core enrichment dependencies are partially unavailable.

#### Scenario: Core list data is available but author enrichment fails

- Given a feed-list request succeeds for core content
- And author enrichment is unavailable
- Then the system or client may render the feed using fallback author data
- And SHALL expose a degraded state rather than failing the entire page

#### Scenario: Viewer-state enrichment fails

- Given core feed items are available
- And viewer-state enrichment fails
- Then the system or client may render default viewer state
- And SHALL keep content browsing available

#### Scenario: Core feed retrieval fails

- Given the system cannot retrieve the core feed list
- When the client requests a feed
- Then the system returns a feed retrieval error

### Requirement: Enforce feed visibility filtering

The system SHALL apply visibility rules before ranking and response shaping.

#### Scenario: Private posts are excluded from unrelated viewers

- Given a post marked as private
- When another viewer requests a feed that could otherwise include the post
- Then the post is excluded

#### Scenario: Followers-only posts require follow relationship

- Given a post marked as followers-only
- When a viewer who does not follow the author requests a feed
- Then the post is excluded

#### Scenario: Mutual-follow posts require mutual relationship

- Given a post marked as mutual-follow-only
- When the viewer and author are not mutual followers
- Then the post is excluded

#### Scenario: Public posts are eligible for latest feed

- Given a post marked as public
- When latest feed retrieval occurs
- Then the post may be included if it satisfies other feed filters

### Requirement: Exclude deleted or invalid content from feeds

The system SHALL exclude content that should not appear in user-facing feed lists.

#### Scenario: Soft-deleted post is filtered out

- Given a soft-deleted post
- When any feed list is built
- Then that post is excluded

#### Scenario: Invalid or moderated content is filtered out

- Given content that should not be shown because of moderation or invalid state
- When any feed list is built
- Then that content is excluded

### Requirement: Apply deterministic V1 ranking

The system SHALL use simple deterministic ranking for latest and following feeds.

#### Scenario: Operator-priority content ranks before ordinary content

- Given a feed containing top or operator-priority content
- When the system sorts the feed
- Then that content ranks ahead of ordinary content

#### Scenario: Remaining V1 content is reverse chronological

- Given non-priority content in the latest or following feed
- When the system sorts the feed
- Then newer content ranks ahead of older content

### Requirement: Preserve a future retrieval-filter-score-mix pipeline

The system SHALL keep the feed architecture compatible with a future retrieval, filtering, scoring, mixing, and pagination pipeline.

#### Scenario: Recommendation work is added later

- Given a future recommendation phase
- When candidate retrieval and scoring are introduced
- Then the feed capability can add those steps without replacing the core contract

### Requirement: Expose a feed list request model

The system SHALL support a list request model that identifies feed type and pagination inputs.

#### Scenario: Client requests latest feed with page parameters

- Given a client requests a feed list
- When the request includes `feedType=latest`, `page`, and `pageSize`
- Then the system applies those inputs to list retrieval

#### Scenario: Client requests following feed with page parameters

- Given a logged-in client requests a feed list
- When the request includes `feedType=following`, `page`, and `pageSize`
- Then the system applies those inputs to following-feed retrieval

### Requirement: Expose a feed list response model

The system SHALL return a list response model that includes feed items and pagination metadata.

#### Scenario: Feed list response includes list and total

- Given a successful feed-list request
- Then the response includes a collection of feed items
- And includes pagination metadata suitable for the current page-based model

### Requirement: Use page-based pagination in V1 and preserve cursor evolution

The system SHALL use page-based pagination in V1 and SHALL record cursor pagination as a later evolution path.

#### Scenario: V1 feed uses page and page size

- Given a V1 client
- When it requests a feed page
- Then the system supports page-number pagination

#### Scenario: Future cursor pagination is introduced

- Given a future product phase
- When cursor pagination is added
- Then it extends the feed contract without invalidating the V1 model

### Requirement: Define legacy endpoint migration behavior

The system SHALL explicitly define how current post-list endpoints relate to the feed capability.

#### Scenario: Existing post-list endpoint remains during transition

- Given the current implementation exposes `/content/post/list`
- When the feed capability is introduced
- Then the team documents whether that endpoint is preserved, wrapped, or superseded

### Requirement: Preserve explicit service boundaries

The system SHALL define feed assembly as an explicit responsibility distinct from raw content storage.

#### Scenario: Content responsibilities remain narrow

- Given content service behavior
- Then content owns post persistence, detail, and content storage
- And does not conceptually own long-term personalized ranking

#### Scenario: Feed assembly owns response shaping

- Given feed list retrieval
- Then feed assembly owns feed-type dispatch, candidate retrieval, filtering, ranking, viewer-state enrichment, and feed-card response assembly

### Requirement: Identify implementation follow-up work

The system SHALL identify concrete backend, frontend, and verification follow-up work before implementation begins.

#### Scenario: Backend follow-up is recorded

- Given the feed capability design
- Then follow-up work identifies backend changes needed to centralize feed assembly and reduce frontend fan-out

#### Scenario: Frontend follow-up is recorded

- Given the feed capability design
- Then follow-up work identifies frontend changes needed for unified feed item rendering

#### Scenario: Verification scenarios are recorded

- Given the feed capability design
- Then follow-up work identifies verification coverage for anonymous latest feed, authenticated following feed, visibility filtering, and degraded dependency behavior
