# Microservices Authentication Project

A comprehensive DevOps learning project demonstrating **authentication persistence across loosely coupled microservices**.

## 🏗 Architecture

This project consists of 5 independent microservices communicating via REST APIs, with shared state managed by Redis and persistent data in PostgreSQL and MongoDB.

### Services

| Service | Tech Stack | Responsibility |
|---------|------------|----------------|
| **Auth Service** | Node.js, Redis | JWT generation, validation, refresh tokens, blacklisting |
| **User Service** | Python, PostgreSQL | User profile management |
| **Product Service** | Go, MongoDB | Product catalog management |
| **Order Service** | Node.js, MongoDB | Order processing, service orchestration |
| **API Gateway** | Go | Single entry point, routing, reverse proxy |

### Infrastructure

- **Redis**: Stores refresh tokens and blacklisted access tokens
- **PostgreSQL**: Stores user profiles
- **MongoDB**: Stores products and orders
- **Docker Compose**: Orchestrates the entire stack

---

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- curl (for testing)

### Run the Project

1. **Start all services:**
   ```bash
   docker compose up -d --build
   ```

2. **Verify services are running:**
   ```bash
   docker ps
   ```

3. **Run the automated test script:**
   ```bash
   chmod +x test-api.sh
   ./test-api.sh
   ```

---

## 🔐 Authentication Flow

1. **Login**: User sends credentials to `Auth Service`.
2. **Token Issue**: Auth Service validates credentials and issues:
   - **Access Token** (JWT, short-lived, 15m)
   - **Refresh Token** (JWT, long-lived, 7d, stored in Redis)
3. **Accessing Resources**: Client sends Access Token in header.
4. **Validation**: Services validate token by:
   - Checking signature (stateless)
   - OR calling Auth Service (stateful check against blacklist)
5. **Refresh**: When Access Token expires, Client sends Refresh Token to get new Access Token.
6. **Logout**: Access Token is blacklisted in Redis; Refresh Token is deleted.

---

## 📚 API Reference

### Auth Service
- `POST /auth/register` - Register new user
- `POST /auth/login` - Login
- `POST /auth/refresh` - Refresh token
- `POST /auth/logout` - Logout

### User Service
- `GET /users/me` - Get profile
- `PUT /users/me` - Update profile

### Product Service
- `GET /products` - List products
- `POST /products` - Create product (Admin only)

### Order Service
- `POST /orders` - Create order
- `GET /orders` - List orders

---

## 🧪 Testing Manually

**1. Register**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"Demo User"}'
```

**2. Login**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

**3. Access Protected Route**
```bash
curl -H "Authorization: Bearer <YOUR_TOKEN>" http://localhost:8080/users/me
```
