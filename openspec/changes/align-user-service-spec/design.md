# Design: Align User Service Spec With Current Implementation

## Overview

The `user` service is already implemented as a layered go-zero flow:

```text
gateway
  -> user-api
      -> user-rpc
          -> MySQL
      -> MinIO (avatar upload path)
frontend
  -> user-api endpoints
```

This change does not alter that shape. It formalizes the behavior that already exists and makes current mismatches explicit.

## Current Capability Model

The current service exposes three groups of behavior.

### 1. Identity and session

- `POST /user/register`
- `POST /user/login`
- `POST /user/token/refresh`

Current design:

- Registration validates username and password at the API layer, then delegates creation to RPC.
- RPC registration creates `user`, `user_profile`, and `user_stat` rows in a single transaction.
- Login verifies password in RPC and generates access and refresh tokens in API.
- Refresh token rotation is handled entirely in API and does not consult RPC.

### 2. Read models

- `GET /user/:id`
- `GET /user/:id/profile`
- `GET /user/:id/stat`
- `POST /user/batch`

Current design:

- API delegates reads to RPC.
- RPC read paths combine DB data and cache where available.
- `GetUserInfo` returns a composed response including basic info plus embedded profile/stat data at RPC level, while API currently exposes separate public endpoints for profile/stat as well.
- Batch lookup is used by frontend aggregation flows such as home feed and comment rendering.

### 3. Authenticated profile mutation

- `POST /user/info`
- `POST /user/avatar`

Current design:

- JWT is enforced at the API route level.
- API derives `userId` from request context rather than trusting client-submitted `userProfile.userId`.
- Avatar upload is handled in API by parsing multipart data, validating image content, uploading to MinIO, then persisting the resulting URL through RPC.
- Profile updates are persisted through RPC and then returned to the client.

## Important Observations

### A. API and RPC are intentionally split

The API layer owns:

- HTTP contract
- JWT token generation and refresh
- multipart avatar handling
- request validation and response shaping

The RPC layer owns:

- persistence
- password verification
- transactional user creation
- cache invalidation
- assembled read models

This split should remain explicit in the spec.

### B. Current behavior is not fully normalized

The current implementation contains several behavior mismatches that should be recorded, not silently normalized:

1. Password rules differ between frontend and backend documentation.
2. `POST /user/info` has historically accepted overly-empty payloads in some flows.
3. `POST /user/avatar` requires multipart upload with an actual file, while some frontend assumptions previously drifted from that contract.
4. API code reads `userId` from raw context values instead of a shared `authctx` helper, unlike some other services in the repository.

### C. Batch lookup is a system dependency, not a convenience endpoint

`POST /user/batch` is required by current frontend data enrichment flows. It should be treated as a first-class capability rather than an auxiliary endpoint.

## Design Decisions For This Change

### Decision 1: Spec the current external contract first

We will document the current externally visible behavior before trying to redesign it.

Reason:

- The project currently lacks a stable source of truth.
- Stabilizing terminology and expectations is the fastest way to reduce confusion.

### Decision 2: Record mismatches as follow-up work

We will not rewrite the current contract in the spec to look cleaner than the code actually is.

Reason:

- That would create a second mismatch, this time between OpenSpec and the repository.
- The spec should distinguish current truth from desired future cleanup.

### Decision 3: Keep one capability spec for the user service

The user service should be represented as one capability spec covering identity, profile, avatar, and stat retrieval.

Reason:

- These behaviors share the same public service boundary.
- Frontend flows treat them as one coherent capability.
- Splitting too early would make current review harder.

## Risks

### Risk 1: Spec captures accidental behavior as intentional

Mitigation:

- Call out risky or undesirable behavior explicitly in the spec and proposal.
- Mark follow-up cleanup in tasks rather than hiding it.

### Risk 2: Future service changes may need more granular specs

Mitigation:

- Start with one user-service capability now.
- Split later only if ownership or product scope meaningfully separates.

### Risk 3: Token lifecycle is only partially documented in code contracts

Mitigation:

- The spec should state that login and refresh are API-issued token flows backed by configured secrets and expiries.
- Any future session revocation or refresh-token storage work should be a separate change.

## Verification Approach

This documentation change is complete when:

- the capability spec reflects the current endpoint set and auth requirements,
- the design explains current layering and dependencies,
- the tasks identify concrete follow-up implementation work without changing code in this change.
