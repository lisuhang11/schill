# Proposal: Align User Service Spec With Current Implementation

## Why

The repository already contains a working `user` service across API, RPC, database schema, frontend integration, and gateway routing, but the behavior is not captured in OpenSpec yet.

Right now, the real contract is split across:

- `service/user/api/user.api`
- `service/user/rpc/user.proto`
- `service/user/api/internal/logic/*`
- `service/user/rpc/internal/logic/*`
- `web/src/api/user.ts`
- `web/前端开发完整文档.md`
- `web/后端接口文档.md`

That makes it hard to answer basic questions consistently:

- What is the intended behavior versus accidental behavior?
- Which fields are part of the stable contract?
- Which mismatches are known and should be corrected later?
- Which user-facing flows depend on the service today?

This change establishes OpenSpec coverage for the current `user` service before any implementation changes are made.

## What Changes

This change does not implement product features. It documents and aligns the current user-service capability by:

1. Creating a capability spec for user identity, profile, avatar, and stats behavior.
2. Capturing the current end-to-end architecture from gateway to API to RPC to persistence.
3. Recording current known contract mismatches and risky behaviors as explicit design constraints.
4. Defining follow-up implementation tasks for stabilizing the service contract.

## Scope

Included:

- Registration
- Login
- Refresh token
- Get user basic info
- Batch get user basic info
- Get user profile
- Get user stat
- Update user profile
- Update avatar
- Current frontend dependencies on the user service

Not included:

- Refactoring non-user services
- Changing gateway behavior
- Implementing admin operations
- Implementing cross-service follow-up fixes in this change

## Success Criteria

- The repository has a user-service capability spec under `openspec/specs/`.
- The current user-service behavior and boundaries are documented in OpenSpec artifacts.
- Known mismatches are recorded as explicit follow-up work rather than remaining tribal knowledge.
