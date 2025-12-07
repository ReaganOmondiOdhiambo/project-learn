# Authentication Flow & Architecture

This document details the authentication patterns used in the Microservices Auth Project.

## 1. Overview

The system uses **stateless JWTs** for access control and **stateful Refresh Tokens** (stored in Redis) for session management. This hybrid approach offers the scalability of JWTs with the security controls of sessions (revocation).

## 2. Token Types

| Token Type | Format | Expiration | Storage (Client) | Storage (Server) | Purpose |
|------------|--------|------------|------------------|------------------|---------|
| **Access Token** | JWT | Short (15m) | Memory / LocalStorage | None (Stateless) | Accessing protected resources |
| **Refresh Token** | JWT | Long (7d) | HttpOnly Cookie / Secure Storage | Redis | Obtaining new Access Tokens |

## 3. Detailed Flows

### A. Registration Flow
1. **Client** sends `POST /auth/register` with email/password.
2. **Auth Service**:
   - Hashes password with bcrypt.
   - Stores user credentials (simulated in Redis for this demo).
   - (Optional) Calls User Service to create profile.
3. **Response**: 201 Created.

### B. Login Flow
1. **Client** sends `POST /auth/login` with credentials.
2. **Auth Service**:
   - Validates credentials.
   - Generates `Access Token` and `Refresh Token`.
   - Stores `Refresh Token` in Redis with 7d expiration (`setEx`).
3. **Response**: Returns both tokens.

### C. Accessing Protected Resources (e.g., User Profile)
1. **Client** sends request to **API Gateway** (`GET /users/me`) with `Authorization: Bearer <ACCESS_TOKEN>`.
2. **API Gateway**:
   - Proxies request to **User Service**.
3. **User Service**:
   - Extracts token.
   - **Option A (Fast)**: Verifies JWT signature locally using shared secret.
   - **Option B (Secure)**: Calls **Auth Service** (`POST /auth/validate`) to check if token is blacklisted.
   - *We implemented Option B to demonstrate service-to-service validation.*
4. **Auth Service**:
   - Checks Redis blacklist.
   - Verifies signature.
   - Returns user payload.
5. **User Service**:
   - Uses user ID from payload to fetch data from PostgreSQL.
   - Returns response.

### D. Token Refresh Flow
1. **Client** detects Access Token expired (401).
2. **Client** sends `POST /auth/refresh` with `Refresh Token`.
3. **Auth Service**:
   - Verifies Refresh Token signature.
   - Checks if Refresh Token exists in Redis whitelist.
   - Issues **new Access Token**.
4. **Response**: Returns new Access Token.

### E. Logout Flow
1. **Client** sends `POST /auth/logout` with Access Token.
2. **Auth Service**:
   - Decodes Access Token to get expiration.
   - Adds Access Token to **Redis Blacklist** until expiration.
   - Deletes Refresh Token from Redis.
3. **Result**: Both tokens are now invalid.

## 4. Service-to-Service Authentication

When **Order Service** needs to call **Product Service**:
1. **Order Service** receives request with User's Access Token.
2. **Order Service** validates token with Auth Service.
3. **Order Service** passes the *same token* (or generates a service token) to **Product Service**.
4. **Product Service** validates the token.

This ensures that the user's identity and permissions are propagated through the call chain.

## 5. Security Considerations

- **HTTPS**: In production, all traffic must be encrypted.
- **Secrets**: JWT secrets and DB passwords should be managed via Kubernetes Secrets or Vault.
- **Rate Limiting**: API Gateway implements rate limiting to prevent abuse.
- **Token Storage**: Clients should store Refresh Tokens in HttpOnly cookies to prevent XSS.
