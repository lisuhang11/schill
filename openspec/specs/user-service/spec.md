# User Service

## Purpose

The user service provides account creation, authentication entry points, user read models, profile updates, avatar updates, and batch user lookups used by other product flows.

## Requirements

### Requirement: Register users

The system SHALL allow anonymous clients to register a new user with username and password.

#### Scenario: Successful registration

- Given a username that does not already exist
- And a password that satisfies backend validation
- When the client calls `POST /user/register`
- Then the service returns success
- And the response includes the new `userId`
- And the system creates the base user record, profile record, and stat record

#### Scenario: Duplicate username

- Given a username that already exists
- When the client calls `POST /user/register`
- Then the service returns a business error indicating the username already exists

### Requirement: Authenticate users and issue tokens

The system SHALL allow anonymous clients to log in with username and password and receive both access and refresh tokens.

#### Scenario: Successful login

- Given a user with valid credentials and normal account status
- When the client calls `POST /user/login`
- Then the service returns success
- And the response includes `userId`
- And the response includes `accessToken` and `refreshToken`
- And the response includes configured expiry values for both tokens

#### Scenario: Invalid credentials

- Given an unknown username or incorrect password
- When the client calls `POST /user/login`
- Then the service returns a business error for invalid credentials

#### Scenario: Abnormal account status

- Given a user whose status is not normal
- When the client calls `POST /user/login`
- Then the service rejects the login

### Requirement: Refresh tokens

The system SHALL allow anonymous clients to exchange a valid refresh token for a new access token and refresh token.

#### Scenario: Successful refresh

- Given a valid refresh token
- When the client calls `POST /user/token/refresh`
- Then the service returns success
- And the response includes a newly generated access token
- And the response includes a newly generated refresh token

#### Scenario: Invalid refresh token

- Given an invalid or expired refresh token
- When the client calls `POST /user/token/refresh`
- Then the service returns a refresh-token error

### Requirement: Read user basic information

The system SHALL expose public user basic information by user ID.

#### Scenario: Get user info

- Given an existing user ID
- When the client calls `GET /user/:id`
- Then the service returns success
- And the response includes basic user fields such as `id`, `username`, `avatar`, `status`, and `createdAt`

### Requirement: Read user profile information

The system SHALL expose public user profile information by user ID.

#### Scenario: Get user profile

- Given an existing user ID
- When the client calls `GET /user/:id/profile`
- Then the service returns success
- And the response includes profile fields such as `gender`, `birthday`, `signature`, `location`, `website`, `company`, `jobTitle`, and `education`

### Requirement: Read user statistics

The system SHALL expose user statistics by user ID.

#### Scenario: Get user stat

- Given an existing user ID
- When the client calls `GET /user/:id/stat`
- Then the service returns success
- And the response includes counts such as `postCount`, `commentCount`, `followerCount`, `followingCount`, `likeCount`, and `collectionCount`

### Requirement: Batch read user basic information

The system SHALL support batch lookup of user basic information for dependent product flows.

#### Scenario: Batch get user info

- Given one or more user IDs
- When the client calls `POST /user/batch`
- Then the service returns success
- And the response includes basic user information keyed by user ID

#### Scenario: Frontend enrichment dependency

- Given feed or comment rendering that requires author information for multiple IDs
- When the frontend collects missing user IDs
- Then it may rely on `POST /user/batch` as an enrichment endpoint

### Requirement: Update profile using authenticated identity

The system SHALL require authentication for profile mutation and SHALL derive the target user identity from the authenticated request context.

#### Scenario: Authenticated profile update

- Given a valid access token
- When the client calls `POST /user/info`
- Then the service uses the authenticated `userId`
- And the service does not trust a client-supplied profile user ID as the source of truth
- And the service persists the profile changes through RPC

#### Scenario: Unauthenticated profile update

- Given no valid access token
- When the client calls `POST /user/info`
- Then the service rejects the request as unauthorized

### Requirement: Update avatar using multipart image upload

The system SHALL require authentication for avatar mutation and SHALL accept avatar content as multipart form upload.

#### Scenario: Successful avatar upload

- Given a valid access token
- And a multipart request containing an avatar image file
- When the client calls `POST /user/avatar`
- Then the API uploads the image to object storage
- And the API persists the resulting avatar URL through RPC
- And the response returns the persisted avatar URL

#### Scenario: Missing avatar file

- Given an authenticated request without an avatar file
- When the client calls `POST /user/avatar`
- Then the service rejects the request as invalid

#### Scenario: Non-image avatar file

- Given an authenticated request with a non-image file
- When the client calls `POST /user/avatar`
- Then the service rejects the request as invalid

## Known Gaps And Constraints

- Frontend and backend password-length expectations are not fully aligned today.
- Avatar upload must use real multipart file content; empty `FormData` is not a valid contract.
- The current codebase has historically had risky behavior around overly-empty profile update payloads, so future contract tightening should be handled explicitly.
- Some API logic currently reads `userId` directly from raw request context values instead of a shared auth helper. This is an internal consistency issue, not a public contract change by itself.
