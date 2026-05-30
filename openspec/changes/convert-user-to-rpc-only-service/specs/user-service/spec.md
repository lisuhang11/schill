# User Service

## MODIFIED Requirements

### Requirement: Authenticate users and issue tokens

The system SHALL allow clients to authenticate through user-rpc and receive both access and refresh tokens issued by user-rpc.

#### Scenario: Successful RPC login

- Given a user with valid credentials and normal account status
- When a caller invokes `UserCenter.Login`
- Then user-rpc returns success
- And the response includes `userId`
- And the response includes `accessToken` and `refreshToken`
- And the response includes configured expiry values for both tokens

#### Scenario: Invalid credentials

- Given an unknown username or incorrect password
- When a caller invokes `UserCenter.Login`
- Then user-rpc returns a business error for invalid credentials

#### Scenario: Abnormal account status

- Given a user whose status is not normal
- When a caller invokes `UserCenter.Login`
- Then user-rpc rejects the login

### Requirement: Refresh tokens

The system SHALL allow clients to exchange a valid refresh token through user-rpc for a new access token and refresh token.

#### Scenario: Successful RPC refresh

- Given a valid refresh token for an existing normal user
- When a caller invokes `UserCenter.RefreshToken`
- Then user-rpc returns success
- And the response includes `userId`
- And the response includes a newly generated access token
- And the response includes a newly generated refresh token

#### Scenario: Invalid refresh token

- Given an invalid or expired refresh token
- When a caller invokes `UserCenter.RefreshToken`
- Then user-rpc returns a refresh-token error

#### Scenario: Refresh for deleted or abnormal user

- Given a valid refresh token whose user no longer exists or is not normal
- When a caller invokes `UserCenter.RefreshToken`
- Then user-rpc rejects the refresh

### Requirement: Update profile using authenticated identity

The system SHALL require authenticated user identity for profile mutation and SHALL receive that identity as part of the RPC request.

#### Scenario: Authenticated RPC profile update

- Given a caller has authenticated a user
- When the caller invokes `UserCenter.UpdateUserProfileInfo`
- Then the RPC request includes the authenticated `userId`
- And user-rpc uses that authenticated `userId` as the target identity
- And user-rpc does not trust a client-supplied profile user ID as the source of truth
- And user-rpc persists the profile changes

### Requirement: Update avatar URL

The system SHALL update user avatars in user-rpc by persisting an already-created avatar URL.

#### Scenario: Successful avatar URL update

- Given a caller has authenticated a user
- And the caller has already converted uploaded avatar content into an avatar URL
- When the caller invokes the user-rpc avatar update method with `userId` and `avatarUrl`
- Then user-rpc persists the avatar URL
- And user-rpc invalidates affected user caches
- And user-rpc returns the persisted avatar URL

#### Scenario: Missing avatar URL

- Given an authenticated user ID
- When the caller invokes the user-rpc avatar update method without an avatar URL
- Then user-rpc rejects the request as invalid

## ADDED Requirements

### Requirement: User-rpc is the authoritative user service boundary

The system SHALL expose complete user business capability through user-rpc rather than depending on a dedicated user-api runtime.

#### Scenario: User capability is available through RPC

- Given another backend service needs user information or user identity behavior
- When it calls user service capability
- Then it SHALL call user-rpc
- And it SHALL NOT depend on `service/user/api` for user business logic

#### Scenario: HTTP layer acts as adapter

- Given an HTTP-facing layer exposes user endpoints
- When it handles user identity, read, or mutation requests
- Then it SHALL adapt the HTTP request to user-rpc
- And it SHALL NOT independently issue user access or refresh tokens

### Requirement: Token signing configuration belongs to user-rpc

The system SHALL store user token signing and expiry configuration with user-rpc.

#### Scenario: Login token configuration

- Given user-rpc starts
- When it loads configuration
- Then it has access-token secret and expiry values
- And it has refresh-token secret and expiry values

#### Scenario: Gateway validates user-rpc tokens

- Given user-rpc issues an access token
- When the gateway validates that token for protected routes
- Then the gateway uses a compatible access-token secret
- And the token is accepted if it is otherwise valid
