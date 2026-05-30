# Design: Patch User Profile With Proto3 Optional Fields

## Overview

The current profile update path is:

```text
client JSON
   -> user-api UpdateUserProfileInfoReq
      -> user-api logic mapping
         -> user-rpc UpdateUserProfileInfoReq
            -> DB update map
```

The problem is not transport itself. The problem is loss of intent.

Today, many fields are ordinary scalar or string values:

- omitted field
- empty string field
- zero scalar field

These can collapse into the same effective shape, so the server cannot reliably distinguish:

```text
do not touch this field
vs
clear this field
vs
set this field to a new value
```

This change introduces PATCH semantics and presence-aware fields so the server can distinguish intent.

## Desired Semantics

For every clearable field, the request must support the following matrix:

```text
Field state in request          Meaning
-----------------------------   --------------------------------
field omitted                   leave existing value unchanged
field present with value        update field to provided value
field present with empty value  clear stored value
```

For example:

```json
{}
```

means no profile changes.

```json
{
  "userProfile": {
    "signature": "hello"
  }
}
```

means set signature to `"hello"`.

```json
{
  "userProfile": {
    "signature": ""
  }
}
```

means clear signature.

## Why Proto3 Optional

`proto3 optional` is the right fit here because it restores field presence for scalar and string fields without requiring a larger protocol redesign.

That gives the RPC layer a way to know:

- whether the client intentionally sent the field,
- and what value the client sent if present.

This is the core requirement for PATCH semantics.

## Fields That Should Be Clearable

The following profile fields should support explicit clearing:

- `birthday`
- `signature`
- `location`
- `website`
- `company`
- `jobTitle`
- `education`

`gender` is different and should be handled deliberately.

### Gender discussion

There are two plausible interpretations:

1. `gender=0` means a real stored domain value of "unknown"
2. omitted `gender` means no change

That means `gender` does not need "empty string means clear" behavior, but it still benefits from presence tracking if we want:

- omitted `gender` => no change
- present `gender=0` => set to unknown
- present `gender=1/2` => set to specific value

So `gender` is presence-sensitive, even though it is not "clearable" in the same way as string/date fields.

## API Contract Shape

The endpoint should be treated as PATCH semantics even if the transport remains `POST /user/info`.

That means:

- omitted fields are ignored,
- provided fields are applied,
- provided empty strings on clearable fields clear the stored value,
- the authenticated `userId` remains server-derived.

This change does not require the route path to become HTTP `PATCH`, but it does require the semantics to become patch-like.

## Mapping Strategy

### API layer

The API layer must preserve whether each field was present in the incoming JSON.

This is the first design fork:

```text
JSON request
   |
   +--> ordinary Go struct fields          -> presence lost
   |
   +--> pointer/presence-aware API fields  -> presence preserved
```

To support proto presence, the API request model must be presence-aware for all patchable fields.

### RPC layer

The RPC request should use `proto3 optional` for all fields whose presence matters.

That includes:

- all clearable string/date-like fields
- `gender` if we want omitted versus explicit `0` to remain distinguishable

### Persistence layer

The RPC update logic should build the DB update map from field presence, not from field truthiness.

That means:

- if field not present: do not include it in updates
- if field present and empty: write empty or null according to storage rules
- if field present and valued: write provided value

## Birthday Handling

`birthday` needs special care because it is stored as a date-like field.

PATCH semantics should be:

```text
birthday omitted      -> no change
birthday = "YYYY-MM-DD" -> set parsed date
birthday = ""         -> clear stored birthday
```

This means the implementation must support explicit nulling of the persisted birthday field rather than only parsing non-empty values.

## Validation Changes

Current validation logic rejects payloads that have "no meaningful change" based on non-empty values. That rule conflicts with PATCH clearing.

The new validation model should instead be:

- reject when no patchable field is present at all,
- accept empty values for clearable fields when those fields are explicitly present,
- validate format/range only when a field is present and not intentionally empty in ways the contract forbids.

## Response Strategy

The response can continue returning the resulting full normalized profile object after applying the patch.

That is a good fit for frontend state refresh because it gives the client the authoritative post-update result.

## Risks

### Risk 1: API layer still loses presence before reaching protobuf

If the API request DTO remains non-pointer regular fields, `proto3 optional` alone will not solve the problem.

Mitigation:

- require presence-aware API DTOs as part of implementation.

### Risk 2: Empty string versus null semantics become inconsistent

If some fields clear to empty string while others clear to database null arbitrarily, the contract becomes confusing.

Mitigation:

- define field-by-field persistence expectations during implementation.
- keep public behavior focused on "field is cleared" rather than leaking storage representation.

### Risk 3: Existing clients may rely on current no-op behavior

Some existing clients may assume empty values are ignored.

Mitigation:

- treat this as an intentional contract change.
- update frontend request construction together with backend implementation.

## Verification Approach

The implementation guided by this design should be considered correct when:

- omitted fields remain unchanged,
- present non-empty fields are updated,
- present empty values clear clearable fields,
- explicit `gender=0` can be distinguished from omitted `gender` if that field is included in the optional set,
- the post-update response matches persisted state.
