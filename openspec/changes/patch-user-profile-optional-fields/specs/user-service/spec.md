# User Service

## MODIFIED Requirements

### Requirement: Update profile using authenticated identity

The system SHALL require authentication for profile mutation, SHALL derive the target user identity from the authenticated request context, and SHALL apply profile changes using PATCH semantics.

#### Scenario: Omitted fields remain unchanged

- Given a valid access token
- And an existing user profile with stored values
- When the client calls `POST /user/info`
- And a patchable field is omitted from `userProfile`
- Then the service SHALL leave that stored field unchanged

#### Scenario: Present non-empty field updates stored value

- Given a valid access token
- When the client calls `POST /user/info`
- And `userProfile.signature` is present with a non-empty value
- Then the service SHALL update the stored signature to that value

#### Scenario: Present empty clearable field clears stored value

- Given a valid access token
- And the profile currently has a stored `signature`
- When the client calls `POST /user/info`
- And `userProfile.signature` is present with value `""`
- Then the service SHALL clear the stored signature

#### Scenario: Birthday can be cleared explicitly

- Given a valid access token
- And the profile currently has a stored birthday
- When the client calls `POST /user/info`
- And `userProfile.birthday` is present with value `""`
- Then the service SHALL clear the stored birthday

#### Scenario: Explicit unknown gender is distinguishable from omitted gender

- Given a valid access token
- When the client calls `POST /user/info`
- And `userProfile.gender` is explicitly present with value `0`
- Then the service SHALL treat that as an intentional update to unknown gender
- And SHALL NOT treat it as equivalent to omission

#### Scenario: User identity remains server-derived

- Given a valid access token
- When the client calls `POST /user/info`
- Then the service SHALL use the authenticated `userId`
- And SHALL NOT trust a client-supplied profile user ID as the source of truth

#### Scenario: No patch fields present is invalid

- Given a valid access token
- When the client calls `POST /user/info`
- And no patchable profile fields are present in the request
- Then the service SHALL reject the request as invalid

## ADDED Requirements

### Requirement: Presence-sensitive RPC profile updates

The system SHALL preserve field presence for all presence-sensitive profile update fields across the API and RPC boundary.

#### Scenario: Clearable fields use proto optional presence

- Given a profile update request containing clearable fields
- When the API maps the request to RPC
- Then the RPC request SHALL preserve whether each clearable field was omitted or explicitly provided

#### Scenario: RPC update logic uses presence instead of truthiness

- Given a presence-sensitive RPC profile update request
- When the RPC layer builds database updates
- Then it SHALL decide whether to update a field based on field presence
- And SHALL NOT use non-empty truthiness as the update condition
