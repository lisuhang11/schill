# Design: Build Enterprise Frontend Pages

## Overview

The implementation should create a real product frontend, not a static mockup. The frontend will live in `web/` and use Next.js App Router with TypeScript.

The architecture should separate concerns this way:

```text
app routes
  -> Server Components for initial data
  -> Client Components for interaction
  -> feature components
  -> hooks
  -> typed API clients
  -> shared contract types
  -> backend HTTP APIs / gateway
```

The implementation must start from backend contracts. The known backend inputs include:

- `service/feed/rpc/feed.proto` for feed request, response, feed item, author, and viewer-state shapes.
- `service/search/rpc/search.proto` for search endpoints and response models.
- `docs/swagger/*.json` for generated HTTP contracts for content, comment, interaction, and relation services.
- other service `.api` files or generated HTTP code if present at implementation time.
- `service/*/rpc/*.proto` for secondary type context when HTTP documentation is incomplete.
- `db.sql` for enum and persistence-shape clarification when API documentation exposes only primitive numbers.

## Current State

`web/` currently contains:

- `frontend-style-guide.md`
- Codex UI skill files

It does not currently contain a Next.js app skeleton, package manifest, `app/` routes, shared components, or API client code. Therefore the first implementation step should be to create the frontend foundation before page work.

The backend already contains a feed RPC contract:

```text
GetFeedListReq
  feedType
  page
  pageSize
  currentUserId

GetFeedListResp
  code
  msg
  total
  list: FeedItem[]
```

The backend search HTTP API exposes:

- `GET /search/post`
- `GET /search/user`
- `GET /search/topic`

with `keyword`, `page`, and `pageSize` query parameters and paginated response payloads.

Generated Swagger HTTP contracts also exist for:

- content post list, detail, create, update, delete, topic list, view count, top, and essence operations,
- comment list, replies, create, delete, and vote operations,
- post star, collection, share, and batch viewer-state operations,
- relation follow, unfollow, follow status, follower list, and following list operations.

The gateway exposes `/api` HTTP routes that adapt to RPC services, including auth and feed routes. When generated Swagger is missing for gateway routes, `service/gateway/internal/handler/routes.go` and `service/gateway/internal/types/types.go` are the practical source of truth.

## Frontend Architecture

### App Router Layout

Use the App Router layout model:

- `app/layout.tsx` for root document structure and global styles.
- `app/page.tsx` for the primary feed/home page.
- route-level `loading.tsx` and `error.tsx` where useful.
- feature folders for route-local UI when it keeps ownership clear.

The exact folder names should follow `web/frontend-style-guide.md`. If the guide conflicts with standard App Router requirements, keep Next.js-required files in their required locations and apply the guide to feature and shared code organization.

### Server Component Responsibilities

Server Components should:

- fetch initial feed/search data using typed API helpers,
- normalize backend errors into renderable error states,
- pass serializable data into Client Components,
- avoid browser-only APIs.

### Client Component Responsibilities

Client Components should:

- handle pagination, tabs, filters, retry, refresh, and form submission,
- use React Hook Form and Zod for form state and validation,
- render optimistic or pending UI only where behavior is clear,
- preserve accessible controls and keyboard-friendly interactions.

### API Client Layer

Create a typed API layer that:

- centralizes `fetch` configuration,
- reads `NEXT_PUBLIC_API_BASE_URL` or a server-only equivalent from environment configuration,
- validates request construction with typed inputs,
- handles backend `{ code, msg, ... }` style responses consistently,
- exposes narrow functions such as `getFeedList`, `searchPosts`, `searchUsers`, and `searchTopics`.

Do not scatter raw `fetch` calls across page components.

### Contract Source Priority

Use this priority when deriving frontend behavior:

1. gateway route/type code under `service/gateway/internal/handler/routes.go` and `service/gateway/internal/types/types.go`,
2. generated Swagger JSON under `docs/swagger/` for service-level HTTP route, method, parameter, body, and response shape,
3. `.proto` files where present, especially `service/search/rpc/search.proto`,
4. generated or handwritten HTTP API code if Swagger and `.api` disagree,
5. `.proto` files for RPC-only type context,
6. `db.sql` for enum meanings and storage details when the API exposes only numeric or string primitives.

If a page requires a capability that has only an RPC contract and no confirmed gateway route, keep that capability behind an isolated API client boundary and document it as an integration gap.

## Type Model

Create `types.ts` files for shared contracts. The feed type model should map the RPC contract closely:

- `FeedListRequest`
- `FeedListResponse`
- `FeedItem`
- `FeedAuthor`
- `FeedViewerState`
- `FeedType`

Search types should map `service/search/rpc/search.proto`:

- `SearchPostRequest`
- `SearchPostResponse`
- `SearchPostItem`
- `SearchUserRequest`
- `SearchUserResponse`
- `SearchUserItem`
- `SearchTopicRequest`
- `SearchTopicResponse`
- `SearchTopicItem`

Use string literal unions only when the backend contract makes the valid values clear. Otherwise use a conservative string or number type and document the known values near the type.

Post publishing and editing should include these frontend-level types:

- `PostVisibility = 0 | 10 | 20 | 50 | 90`
- `PostContentType = 2` for the first frontend version
- `PostContentItem` with `type: 2`, `content: string`, and `sort: number`
- `CreatePostInput` and `UpdatePostInput` matching the HTTP request body but with form helpers for comma-separated tags

Visibility labels are sourced from `db.sql`:

- `0`: private
- `10`: paid or charge-gated
- `20`: followers-only
- `50`: mutual-follow-only
- `90`: public

Although `db.sql` lists post content types `1` through `8`, this change only exposes plain text content with backend type `2`. Markdown, images, code blocks, video, audio, links, attachments, and paid resources are deferred.

## Page Design

The UI should follow the marine blue-white style guide:

- light background with white content surfaces,
- restrained blue as the primary action color,
- small warm or mint accents for emphasis,
- clear cards and lists with stable spacing,
- no decorative-heavy landing page before the usable product,
- mobile-safe layout and text wrapping.

The primary page should expose the usable feed/search experience immediately. It should not be a marketing landing page.

## Data States

Every data-driven surface must implement:

- loading state for initial and interactive fetches,
- empty state for successful responses with no records,
- error state for failed requests or backend non-success codes,
- degraded state where non-critical fields such as avatar, cover, summary, or viewer-state flags are missing.

## Form Design

Forms must use:

- React Hook Form for state,
- Zod for validation,
- typed submit payloads,
- inline validation messages,
- disabled or pending states during submit/fetch.

For the first implementation, a search/filter form is acceptable if no create/update HTTP API is available. If a documented create/update HTTP API exists, the implementation may include a post form only when it can strictly match the backend request contract.

For post create/edit forms:

- `contents[]` supports only plain text for now and submits each text block with `type: 2`.
- `tags` is entered as comma-separated text and submitted as the backend string field.
- `topics` is entered as a simple string-array field without autocomplete.
- `visibility` is selected from the confirmed enum values `0`, `10`, `20`, `50`, and `90`.
- Code block, Markdown, image, and rich-media editor controls must not be shown until backend and product scope explicitly include them.

## Backend Contract Policy

Implementation must not invent API behavior.

When connecting a page:

1. Read the HTTP `.api` file or generated API route code.
2. Match query/body fields, response fields, auth requirements, and error-code shape.
3. If an RPC-only capability is needed, read the `.proto` for type context and then verify the HTTP gateway route.
4. If the gateway route is absent, keep the frontend API function isolated, document the missing route in `web/README.md`, and avoid pretending it is fully integrated.

Known gaps to verify and record during implementation:

- Gateway generated Swagger is absent, so gateway route/type code should be cited in README until `docs/swagger/gateway.json` exists.
- Collection-center pages require a collection-list endpoint; current gateway routes document collect/uncollect but not a paginated "my collections" list.

## README Requirements

`web/README.md` should include:

- technology stack,
- environment variables,
- backend endpoint mapping,
- how request and response types were derived,
- run commands,
- test/lint/typecheck commands,
- notes for adding new backend endpoints,
- known integration gaps or assumptions.

## Verification Approach

Implementation is complete when:

- dependencies install successfully,
- the Next.js app builds or typechecks,
- linting passes where configured,
- the implemented page renders loading, empty, and error states,
- React Hook Form + Zod validation can be exercised,
- `README.md` is accurate for the implemented scripts and endpoints.

If the environment cannot install dependencies because of network restrictions, implementation should still produce the files and report the blocked verification command exactly.

## Risks

### Risk 1: Backend HTTP routes are incomplete

Mitigation:

- implement against confirmed routes first,
- isolate unconfirmed feed integration behind a typed API boundary,
- document missing gateway routes instead of faking them.

### Risk 2: Creating too much frontend surface at once

Mitigation:

- prioritize a complete, polished feed/search experience over broad but shallow pages,
- keep components reusable only where there is real repeated behavior.

### Risk 3: Type drift from backend contracts

Mitigation:

- keep contract-derived `types.ts` close to API clients,
- reference source backend files in comments or README,
- add typecheck and tests for request builders and response handling.

### Risk 4: Style guide encoding issues

Mitigation:

- follow the observable palette and layout guidance from the file,
- avoid relying on unreadable text fragments,
- keep visual decisions conservative and product-focused.
