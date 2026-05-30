# Tasks: Align User Service Spec With Current Implementation

## 1. Document current capability

- [ ] Create `openspec/specs/user-service/spec.md`.
- [ ] Capture public endpoints, auth requirements, and response responsibilities.
- [ ] Record frontend dependencies on `/user/batch`, `/user/info`, and `/user/avatar`.

## 2. Validate design against code

- [ ] Confirm API responsibilities from `service/user/api/*`.
- [ ] Confirm RPC responsibilities from `service/user/rpc/*`.
- [ ] Confirm persistence expectations from `db.sql` and user RPC models.

## 3. Record known mismatches

- [ ] Document password-rule mismatch between frontend and backend expectations.
- [ ] Document avatar upload contract requirements.
- [ ] Document risky empty-body or overly-permissive profile update behavior history.
- [ ] Document the current direct context access pattern for `userId`.

## 4. Prepare follow-up implementation candidates

- [ ] Identify which mismatches should become separate implementation changes.
- [ ] Distinguish contract fixes from internal refactors.
- [ ] Leave this change implementation-free and ready for a later `/opsx:apply`.
