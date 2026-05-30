# Tasks: Patch User Profile With Proto3 Optional Fields

## 1. Define the patch contract

- [x] Confirm the endpoint keeps path `POST /user/info` while adopting PATCH semantics.
- [x] Enumerate all fields that must support explicit clearing.
- [x] Decide whether `gender` also uses presence-aware optional semantics.

## 2. Update request models

- [x] Change API request DTOs so patchable field presence is preserved from JSON parsing.
- [x] Change `service/user/rpc/user.proto` so all presence-sensitive fields use `proto3 optional`.
- [x] Regenerate protobuf and RPC bindings after the proto update.

## 3. Update validation and mapping

- [x] Replace truthiness-based "meaningful change" checks with presence-based validation.
- [x] Map API request presence accurately into RPC optional fields.
- [x] Ensure authenticated `userId` remains server-derived.

## 4. Update persistence logic

- [x] Build DB updates from field presence rather than non-empty values.
- [x] Support clearing `birthday`.
- [x] Support clearing string profile fields without treating them as no-op.
- [x] Keep cache invalidation behavior correct after partial updates.

## 5. Verify behavior end to end

- [ ] Add tests for omitted field => unchanged.
- [ ] Add tests for present non-empty field => updated.
- [ ] Add tests for present empty clearable field => cleared.
- [x] Add tests for explicit `gender=0` semantics if included in the optional set.
- [x] Validate frontend request payload expectations against the new contract.
