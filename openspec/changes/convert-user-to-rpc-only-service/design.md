# Design: Convert User Service To RPC-Only Boundary

## Overview

The target user-service shape is:

```text
frontend / gateway / adapter
  -> user-rpc
      -> MySQL
      -> Redis cache
      -> MQ consumers
```

The user RPC service becomes the owner of both identity verification and token issuance. The removed API layer should not remain the place where user tokens are created or refreshed.

## Current Boundary

Today:

```text
gateway
  -> user-api
      -> user-rpc
          -> MySQL / Redis
      -> MinIO
```

The current split is:

- `user-api` validates HTTP requests, issues JWTs, refreshes JWTs, parses multipart avatar uploads, uploads avatar content to MinIO, and maps HTTP responses.
- `user-rpc` creates users, verifies login credentials, reads user data, updates profile data, updates avatar URLs, invalidates user caches, and consumes user-related events.

This split is useful for HTTP delivery, but it is not the desired long-term service boundary if other services call user capability through RPC.

## Target Boundary

After this change:

```text
callers
  -> user-rpc
      identity/session:
        Register
        Login
        RefreshToken
      read models:
        GetUserInfo
        GetUserProfileInfo
        GetUserStat
        BatchGetUserBasicInfo
      mutations:
        UpdateUserProfileInfo
        UpdateAvatarUrl
        UpdateUserStatus
```

Any HTTP-facing layer becomes an adapter. It may translate HTTP requests into RPC calls, but it must not independently own user token issuance.

## Proto Direction

The RPC contract should make token issuance explicit:

```text
LoginReq
  username
  password

LoginResp
  userId
  accessToken
  accessExpireIn
  refreshToken
  refreshExpireIn

RefreshTokenReq
  refreshToken

RefreshTokenResp
  userId
  accessToken
  accessExpireIn
  refreshToken
  refreshExpireIn
```

The RPC contract should keep user mutations identity-explicit:

```text
UpdateUserProfileInfoReq
  userId
  profilePatch

UpdateAvatarUrlReq
  userId
  avatarUrl
```

RPC callers must pass the authenticated user ID after gateway/auth validation. The user-rpc method must not trust a profile body user ID as the source of truth.

## Token Issuance

`user-rpc` will use the existing `common/jwt` package to generate and parse tokens.

The user-rpc configuration needs JWT fields equivalent to the current user-api configuration:

```text
Jwt:
  AccessSecret
  AccessExpire
  RefreshSecret
  RefreshExpire
```

Login flow:

```text
Login
  -> validate username/password
  -> verify user exists, not deleted, normal status, password matches
  -> update last login/active time
  -> generate access token
  -> generate refresh token
  -> return token pair and expiries
```

Refresh flow:

```text
RefreshToken
  -> parse refresh token with RefreshSecret
  -> reject invalid or expired token
  -> optionally verify user still exists and is normal
  -> generate new access token
  -> generate new refresh token
  -> return token pair and expiries
```

The optional user-status check during refresh is recommended so disabled/deleted users cannot continue refreshing sessions.

## Avatar Boundary

Avatar upload should not move into user-rpc in this change.

Reason:

- gRPC is not the best first boundary for browser multipart uploads.
- Passing large file bytes through user-rpc introduces message-size, streaming, content sniffing, and storage concerns.
- Current user-rpc already has a simple and useful mutation: persist the avatar URL and invalidate user caches.

Target responsibility:

```text
HTTP/media boundary:
  multipart file -> validation -> object storage -> avatarUrl

user-rpc:
  avatarUrl -> user row update -> cache invalidation
```

If the project later needs centralized media handling, that should be a separate media/storage capability rather than hidden inside user-rpc.

## Deployment Impact

Deployment references to `user-api` should be removed as part of this cutover. The existing gateway is an HTTP reverse proxy and cannot directly proxy gRPC, so user HTTP routes should be removed rather than pointed at `user-rpc`.

Affected areas likely include:

- Docker Compose service definitions.
- Kubernetes `app-user.yaml`.
- Gateway route configuration targeting `http://user-api:8888`.
- Build scripts that build `schill-user-api`.
- Frontend/user HTTP documentation, if an adapter path changes.

## Risks

### Risk 1: Frontend loses HTTP user endpoints

Mitigation:

- This cutover intentionally removes the HTTP user API.
- Treat future HTTP exposure as a separate adapter concern if needed.

### Risk 2: RPC token issuance creates config drift with gateway JWT validation

Mitigation:

- Keep gateway access-token validation secret aligned with user-rpc `Jwt.AccessSecret`.
- Document which service owns signing versus verification.

### Risk 3: Refresh-token behavior remains stateless

Mitigation:

- This change keeps the current stateless JWT refresh model.
- Server-side refresh-token revocation or rotation storage should be a separate change if needed.

### Risk 4: Proto regeneration can break existing RPC callers

Mitigation:

- Inventory current user-rpc callers before applying changes.
- Update callers in the same implementation change where proto methods or message names change.
- Prefer additive proto changes where possible before removing old fields or methods.
