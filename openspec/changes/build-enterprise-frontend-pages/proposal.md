# Proposal: Build Enterprise Frontend Pages

## Why

The repository currently has backend service contracts and a frontend style guide, but `web/` does not yet contain a Next.js App Router application that consumes those contracts.

The project needs a production-oriented frontend foundation that can be implemented against the existing backend APIs without guessing request and response shapes. The most important known contract source for the feed experience is `service/feed/rpc/feed.proto`; HTTP-facing contracts must be read from the available `.api` files and generated API code during implementation. Where an HTTP endpoint is missing for an RPC capability, the implementation must surface that integration gap instead of inventing a contract silently.

## What Changes

This change proposes building a TypeScript + React + Next.js App Router frontend under `web/` that:

1. follows `web/frontend-style-guide.md` for structure, naming, Tailwind, component usage, and visual direction,
2. models backend request and response types from the real HTTP API definitions and available `.proto` contracts,
3. uses Server Components for initial data loading and Client Components for interactive behavior,
4. implements loading, empty, and error states for data-driven pages,
5. uses React Hook Form and Zod for forms and validation,
6. includes page code, hooks, API clients, shared types, and a frontend `README.md` for interface wiring, running, and testing.

## Scope

Included:

- Next.js App Router project foundation inside `web/` if no compatible app already exists.
- Enterprise-grade feed/home page implementation using the backend feed contract where an HTTP route exists or can be documented through the gateway.
- Search or list-oriented page integration where HTTP `.api` contracts are already available.
- Article detail, comment, interaction, relation, topic, and post publishing/editing surfaces where Swagger HTTP contracts are available.
- Typed API client layer and shared `types.ts` definitions derived from backend contracts.
- Server Component initial data fetch paths.
- Client Component interactions such as filtering, pagination, retry, refresh, or form submission.
- Loading, empty, error, and degraded data states.
- React Hook Form + Zod form validation for user-entered query/filter/post data used by the implemented pages.
- `web/README.md` explaining backend interface mapping, environment variables, run commands, and test commands.
- A documented integration-gap list for capabilities that only have RPC contracts or unclear HTTP routes.

Not included:

- Changing backend service behavior unless implementation uncovers a blocking frontend integration gap.
- Inventing undocumented HTTP endpoints for RPC-only capabilities.
- Building pages unrelated to currently documented backend capabilities.
- Adding speculative dashboards, admin surfaces, or marketing pages.
- Replacing backend-generated contracts with hand-waved frontend-only models.
- Rich post-content editing beyond plain text, such as images, videos, code blocks, Markdown rendering, audio, links, attachments, or paid resources.

## Assumptions

- The frontend should be implemented in `web/`.
- The target stack is TypeScript, React, Next.js App Router, Tailwind CSS, React Hook Form, and Zod.
- `web/frontend-style-guide.md` is the source of truth for visual and project conventions even though the file is currently displayed with mojibake in this environment.
- Backend HTTP APIs are the primary integration contract. RPC `.proto` files are valid secondary sources when HTTP documentation is incomplete or when planning typed data models.
- If the feed RPC has no HTTP gateway route yet, the frontend implementation should either connect to the actual gateway route discovered in code or document the backend gap in `README.md` and keep the API client boundary ready.
- Post publishing/editing initially supports only plain-text content items using backend content type `2`.
- The tags form field is entered as a comma-separated string because the backend `post.tags` column is stored as comma-separated text.
- The topics form field is a plain string-array input without autocomplete for this change.
- Post visibility values are taken from `db.sql`: `0` private, `10` paid/charge, `20` followers-only, `50` mutual-follow-only, and `90` public.
- If an expected HTTP API is missing, the implementation should record it as a known change gap rather than blocking unrelated confirmed pages.

## Success Criteria

- `web/` contains a runnable Next.js App Router frontend.
- Implemented pages use Server Components for initial fetch and Client Components for interaction.
- Frontend request and response types match backend HTTP contracts and relevant `.proto` structures.
- Loading, empty, error, and form validation states are visible in the implemented flows.
- The implementation includes `.tsx` pages/components, hooks, API clients, and `types.ts`.
- `web/README.md` explains API mapping, environment setup, running, testing, and known backend integration gaps.
- The change is ready for `/opsx:apply` implementation without additional planning.
