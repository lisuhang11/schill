# Tasks: Build Enterprise Frontend Pages

## 1. Confirm contracts and project setup

- [x] Read `web/frontend-style-guide.md` and extract applicable folder, naming, Tailwind, and component conventions.
- [x] Inventory backend HTTP API definitions under `service/*/api`.
- [x] Inventory generated Swagger HTTP contracts under `docs/swagger/*.json`.
- [x] Inventory relevant RPC contracts under `service/*/rpc/*.proto`, starting with `service/feed/rpc/feed.proto`.
- [x] Use `db.sql` to confirm enum meanings that are only numeric in Swagger, including post visibility and content type.
- [x] Identify which frontend pages can be fully integrated through existing HTTP routes.
- [x] Record any RPC-only capability that lacks a confirmed HTTP gateway route.

## 2. Create the Next.js App Router foundation

- [x] Add `web/package.json` with Next.js, React, TypeScript, Tailwind, React Hook Form, Zod, and test/lint scripts.
- [x] Add TypeScript, Next.js, PostCSS, Tailwind, and ESLint configuration consistent with the project style.
- [x] Create `app/layout.tsx`, global styles, and route structure under `web/`.
- [x] Create shared UI primitives only where needed by the implemented pages.
- [x] Verify the project can typecheck or document dependency-install blockers.

## 3. Implement backend contract types and API clients

- [x] Create shared feed types matching `service/feed/rpc/feed.proto`.
- [x] Create search types matching `service/search/rpc/search.proto`.
- [x] Create content, comment, interaction, and relation types from `docs/swagger/*.json`.
- [x] Create post create/edit form types that submit only plain-text `contents[]` items with backend content type `2`.
- [x] Create a post visibility union for `0`, `10`, `20`, `50`, and `90` using labels confirmed from `db.sql`.
- [x] Create a centralized fetch helper that handles base URL, query strings, JSON parsing, and backend error shape.
- [x] Implement typed API functions for confirmed HTTP endpoints.
- [x] Keep unconfirmed RPC-backed integrations isolated and documented.

## 4. Build Server Component data flows

- [x] Implement the initial feed/home route as a Server Component.
- [x] Fetch initial data on the server through typed API helpers.
- [x] Normalize success, empty, and error outcomes into serializable props.
- [x] Add route-level or component-level loading UI for initial render.

## 5. Build Client Component interactions

- [x] Implement tabs, filters, search, pagination, retry, or refresh as Client Components.
- [x] Use React Hook Form and Zod for the search/filter form or a documented backend-backed form.
- [x] Build post create/edit form behavior with plain-text content only.
- [x] Parse tags from comma-separated user input into the backend string contract without adding tag autocomplete.
- [x] Model topics as a simple string-array input without autocomplete.
- [x] Render visibility as a select/radio group using the confirmed enum values.
- [x] Show validation errors inline.
- [x] Render pending states during interactive requests.
- [x] Keep all interactive state inside Client Components.

## 6. Implement enterprise-ready UI states

- [x] Render loading states for initial and interactive data fetching.
- [x] Render empty states for no feed/search results.
- [x] Render error states for failed requests and backend non-success codes.
- [x] Render degraded fallbacks for missing avatars, covers, summaries, tags, or viewer state.
- [x] Verify the layout is responsive and follows the marine blue-white style direction.

## 7. Add README and verification

- [x] Create `web/README.md`.
- [x] Document environment variables and backend base URL setup.
- [x] Document endpoint-to-file contract mapping.
- [x] Document missing or unconfirmed HTTP routes discovered during implementation, especially user, feed, and collection-list gaps.
- [x] Document the post editor limitations: plain text only, comma-separated tags, topics without autocomplete, and confirmed visibility values.
- [x] Document how to run, build, lint, typecheck, and test.
- [x] Document known backend integration gaps and assumptions.
- [x] Run available verification commands and record any blocked commands in the final implementation summary.
