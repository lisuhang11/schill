# Proposal: Patch User Profile With Proto3 Optional Fields

## Why

The current `POST /user/info` update flow cannot cleanly express three different client intents:

1. leave a field unchanged
2. set a field to a non-empty value
3. clear a previously stored value

Today, the request model relies on regular scalar and string fields, which collapses "field omitted" and "field present with zero/empty value" into the same shape at the RPC boundary. That makes profile updates awkward and prevents a reliable PATCH-style contract.

This is especially problematic for user profile fields that should be clearable, such as:

- `birthday`
- `signature`
- `location`
- `website`
- `company`
- `jobTitle`
- `education`

The current logic also uses "meaningful change" heuristics that treat empty values as "no change", which blocks legitimate clear operations.

## What Changes

This change converts the profile update contract to explicit PATCH semantics by:

1. changing the user-profile update endpoint contract from replace-like behavior to PATCH behavior,
2. representing clearable profile fields as `optional` fields in `proto3`,
3. making the API and RPC layers preserve field presence through the request path,
4. defining how omitted, present non-empty, and present empty values are interpreted,
5. updating validation rules so "clear this field" is valid behavior rather than rejected as no-op.

## Scope

Included:

- `POST /user/info` contract redesign to PATCH semantics
- `UpdateUserProfileInfoReq` field presence model
- `proto3 optional` usage for clearable profile fields
- API-to-RPC mapping rules for omitted versus provided fields
- persistence rules for clearable fields
- follow-up expectations for frontend request payloads

Not included:

- changing login, register, or token flows
- changing avatar upload behavior
- introducing JSON Patch or RFC 6902 operations
- redesigning public read endpoints
- implementing nickname support

## Success Criteria

- The change clearly defines PATCH semantics for profile updates.
- The change identifies which fields must support explicit clearing.
- The change specifies how field presence survives from HTTP payload to protobuf to persistence.
- The change is ready for implementation through a later apply step.
