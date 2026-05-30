# Proposal: Add Feed System

## Why

The repository already contains the building blocks of a community content experience:

- posts and post content
- follow relationships
- interaction counters such as likes, collections, and shares
- comment flows
- search and Elasticsearch indexing
- a homepage that already renders `latest` and `following` lists

However, the current behavior is still shaped as a post-listing capability rather than a dedicated feed capability.

Today, feed behavior is spread across:

- `service/content` list APIs and RPC logic
- `service/relation` follow data
- `service/interaction` viewer-state enrichment
- `service/user` batch author lookup
- frontend composition in `web/src/views/HomePage.vue`

That makes it difficult to answer product and architecture questions consistently:

- What kinds of feeds does the product support?
- Which rules determine visibility and ranking?
- What should a feed item contain?
- Which behavior belongs to content, and which belongs to feed assembly?
- How should the current latest/following implementation evolve as scale and product complexity grow?

This change defines a first-class feed capability so the project has a stable source of truth before implementation expands.

## What Changes

This change defines a feed system proposal that:

1. establishes feed as a distinct product capability rather than an incidental post-list endpoint,
2. defines the first supported feed types and their intended behavior,
3. specifies the feed item contract used by homepage and related list views,
4. documents ranking, filtering, visibility, and enrichment responsibilities,
5. records an incremental architecture path from the current implementation to a more scalable feed system.

## Scope

Included:

- homepage feed capability
- latest feed
- following feed
- reusable feed item response model
- pagination and loading model for feed lists
- visibility and viewer-state requirements
- service-boundary guidance for content, relation, interaction, and feed assembly
- phased evolution guidance for future recommendation support

Not included:

- implementing recommendation algorithms
- introducing ads or commercial insertion logic
- building real-time push delivery
- changing comment system behavior
- redesigning search as part of this change
- migrating all existing endpoints in this change

## Success Criteria

- The repository has a proposed OpenSpec change for a feed capability.
- The change defines clear V1 product requirements for latest and following feeds.
- The change defines a stable feed item model that can guide backend and frontend integration.
- The change records an architecture path that starts from the current MySQL-based implementation without forcing premature complexity.
- The change is ready for a later apply step and implementation planning.
