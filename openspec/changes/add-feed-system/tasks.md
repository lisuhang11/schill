# Tasks: Add Feed System

## 1. Define the feed capability and public contract

- [x] Create a feed capability spec under `openspec/specs/` or change-local specs if the team wants implementation-scoped spec work first.
- [x] Define supported V1 feed types: `latest` and `following`.
- [x] Define future-compatible feed-type extensibility for `recommend`, `user_posts`, and `topic_posts`.
- [x] Specify authentication expectations for anonymous and logged-in feed access.

## 2. Define the feed item response model

- [x] Specify the canonical feed item fields for content, author, interaction counters, viewer state, and feed metadata.
- [x] Clarify which fields are mandatory in V1 and which are reserved for future ranked feeds.
- [x] Specify empty-state and degraded-state expectations for partially unavailable enrichment dependencies.

## 3. Define filtering and ranking behavior

- [x] Specify visibility filtering rules for private, followers-only, mutual-follow, and public posts.
- [x] Specify deletion and moderation exclusion behavior for feed lists.
- [x] Specify V1 ranking rules for latest and following feeds.
- [x] Record the future retrieval-filter-score-mix pipeline as a design constraint for later recommendation work.

## 4. Define API and pagination behavior

- [x] Specify the feed list request model, including feed type and pagination inputs.
- [x] Specify the feed list response model, including list payload and pagination metadata.
- [x] Record page-based pagination as the initial model and cursor pagination as a planned evolution path.
- [x] Clarify whether legacy `/content/post/list` will be preserved, wrapped, or superseded by a feed-specific endpoint.

## 5. Plan service-boundary and implementation evolution

- [x] Identify which current responsibilities remain in `content`, `relation`, `interaction`, and `user`.
- [x] Define the target feed-assembly responsibility boundary.
- [x] Decide whether the first implementation stays in the current content service or introduces a separate feed module.
- [x] Record the scale-up path from read fan-out to a possible hybrid distribution model for following feed delivery.

## 6. Prepare implementation follow-up

- [x] Identify backend changes needed to reduce frontend fan-out and centralize feed assembly.
- [x] Identify frontend integration changes needed for unified feed item rendering.
- [x] Identify verification scenarios for anonymous latest feed, authenticated following feed, visibility filtering, and degraded dependency behavior.
- [x] Split implementation work into concrete apply-ready tasks once the team accepts the contract.
