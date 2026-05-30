# Proposal: Convert User Service To RPC-Only Boundary

## Why

The current user service exposes both `user-api` and `user-rpc`. The API layer owns HTTP concerns, JWT token issuance, refresh-token handling, avatar multipart upload, and request/response shaping, while the RPC layer owns persistence and most user business logic.

The next architecture direction is to let RPC services call each other directly and treat `user-rpc` as the complete user capability provider. Keeping a dedicated `user-api` layer for user behavior creates an extra boundary that no longer matches the desired service model.

This change proposes making `user-rpc` the authoritative boundary for user identity, session token issuance, user read models, and user mutations. HTTP exposure can be handled by gateway or another adapter layer, but user business capability should live behind the RPC contract.

## What Changes

This change will:

1. Remove the standalone `service/user/api` runtime from the intended architecture.
2. Redesign `service/user/rpc/user.proto` so user-rpc exposes the complete user service contract.
3. Move access-token and refresh-token issuance from user-api into user-rpc login and refresh flows.
4. Add RPC support for refresh-token exchange.
5. Keep avatar file upload out of user-rpc for now; user-rpc will persist an avatar URL, while an HTTP/media boundary remains responsible for converting uploaded file content into a URL.
6. Update dependent configuration and deployment references so callers use user-rpc instead of user-api where applicable.

## Scope

Included:

- User registration RPC contract and logic alignment.
- User login RPC response including access and refresh tokens.
- Refresh-token RPC method and token rotation behavior.
- User info, profile, stat, and batch basic-info RPC contract alignment.
- Authenticated profile update as an RPC request containing the authenticated user ID.
- Avatar URL update through RPC.
- Removal of user-api deployment/build references as part of the RPC-only cutover.

Not included:

- Implementing a new gateway-to-RPC adapter.
- Designing a generic media or object-storage service.
- Streaming avatar bytes through user-rpc.
- Changing non-user service domain behavior except where their user-rpc client contract must be updated.

## Success Criteria

- `user-rpc` can register users, authenticate users, issue login tokens, refresh tokens, read user data, update profile data, and update avatar URLs through RPC methods.
- Token signing configuration lives with `user-rpc`.
- No intended runtime path depends on `service/user/api` for user business logic.
- Avatar file upload responsibility is explicitly outside user-rpc and the RPC contract only accepts persisted avatar URLs.
- Dependent services that enrich or validate user data continue to call user-rpc successfully.
