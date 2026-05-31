# Schill Web

TypeScript + React + Next.js App Router frontend for the Schill community blog system.

## Stack

- Next.js App Router
- TypeScript
- Tailwind CSS
- React Hook Form
- Zod
- lucide-react icons

## Backend Base URL

Create `.env.local` in `web/` when running locally:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:8086
```

Server-side API helpers also accept `API_BASE_URL`. If neither variable is set, the frontend uses `http://localhost:8086`, matching `service/gateway/etc/gateway.yaml`.

## Implemented Pages

- `/` - homepage using gateway `GET /api/posts` and `GET /api/topics`
- `/login` - login using gateway `POST /api/auth/login`
- `/register` - registration using gateway `POST /api/auth/register`
- `/search` - post/user/topic search using gateway `GET /api/search/post`, `GET /api/search/user`, and `GET /api/search/topic`
- `/topics` - topic list using gateway `GET /api/topics`; topic detail is deferred because no topic-posts HTTP route is confirmed
- `/posts/[postId]` - post detail, plain-text content rendering, comments, vote, star, collect, and share actions
- `/posts/new` - plain-text post creation form using React Hook Form and Zod

## Contract Mapping

HTTP contracts are derived from:

- `service/gateway/internal/handler/routes.go`
- `service/gateway/internal/types/types.go`
- `docs/swagger/content.json`
- `docs/swagger/comment.json`
- `docs/swagger/interaction.json`
- `docs/swagger/relation.json`
- `service/search/api/search.api`

Secondary type context is derived from:

- `service/feed/rpc/feed.proto`
- `service/user/rpc/user.proto`
- other `service/*/rpc/*.proto` files when needed

Gateway wraps RPC responses as `{ code, msg, data }`, so `lib/api.ts` unwraps `data` before returning typed frontend payloads.

Authenticated gateway actions currently use the project-local `X-User-Id` header. The login form stores `schill:userId`, `schill:accessToken`, and `schill:refreshToken` in `localStorage`; API actions that require identity send `X-User-Id` from local storage.

Enum clarification is derived from `db.sql`:

- post visibility: `0` private, `10` paid/charge-gated, `20` followers-only, `50` mutual-follow-only, `90` public
- post content type: the database supports multiple content types, but this frontend currently submits only `2` for plain text

## Post Editor Limits

The first editor intentionally keeps the backend contract narrow:

- `contents[]` supports only one plain-text block with `type: 2`
- tags are entered as comma-separated text and submitted as the backend `tags` string
- topics are entered as comma-separated strings and submitted as `topics: string[]`
- no topic autocomplete
- no Markdown, code block, image, video, audio, attachment, link, or paid-resource editor controls

## Known Integration Gaps

- Gateway has real `/api` routes, but no generated `docs/swagger/gateway.json` is present. The current source of truth for gateway routes is `service/gateway/internal/handler/routes.go`.
- Gateway supports feed and auth routes, so login/register and feed clients use real `/api` paths.
- Gateway supports collect/uncollect and share, but no paginated "my collections" route is documented.

## Commands

```bash
npm install
npm run dev
npm run typecheck
npm run build
npm run lint
```

On Windows PowerShell, if script execution blocks `npm`, use `npm.cmd`:

```powershell
npm.cmd install
npm.cmd run dev
npm.cmd run typecheck
```

## Adding New Endpoints

When the backend adds or changes HTTP routes:

1. Regenerate or update Swagger JSON under `docs/swagger/`.
2. Update `lib/types.ts` from the HTTP response/request contract.
3. Add a narrow function in `lib/api.ts`.
4. Keep raw `fetch` calls out of page components.
5. Update this README if the endpoint resolves a known integration gap.
