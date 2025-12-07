#!/bin/bash

# API TEST SCRIPT
# ===============
# Tests the full authentication flow across microservices

API_URL="http://localhost:8080"
EMAIL="test-$(date +%s)@example.com"
PASSWORD="password123"

echo "🚀 Starting Microservices Auth Test"
echo "==================================="
echo "Target: $API_URL"
echo "Test User: $EMAIL"
echo ""

# 1. Register
echo "1️⃣  Registering User..."
REGISTER_RES=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"name\":\"Test User\"}")

echo "Response: $REGISTER_RES"
echo ""

# 2. Login
echo "2️⃣  Logging In..."
LOGIN_RES=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

# Extract Token
TOKEN=$(echo $LOGIN_RES | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)
REFRESH_TOKEN=$(echo $LOGIN_RES | grep -o '"refreshToken":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Login failed! No token received."
    exit 1
fi

echo "✅ Login successful!"
echo "Access Token: ${TOKEN:0:20}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:20}..."
echo ""

# 3. Access User Profile (User Service)
echo "3️⃣  Accessing User Profile (User Service)..."
PROFILE_RES=$(curl -s -X GET "$API_URL/users/me" \
  -H "Authorization: Bearer $TOKEN")

echo "Response: $PROFILE_RES"
echo ""

# 4. Create Product (Product Service) - Needs Admin, so we'll just list for now
echo "4️⃣  Listing Products (Product Service)..."
PRODUCTS_RES=$(curl -s -X GET "$API_URL/products")
echo "Response: $PRODUCTS_RES"
echo ""

# 5. Create Order (Order Service -> Product Service)
# First create a dummy product directly via internal service if possible, or just try to order a fake one
# Since we can't easily create a product without admin token, we'll skip order creation in this basic test
# or we can try to access orders list
echo "5️⃣  Accessing Orders (Order Service)..."
ORDERS_RES=$(curl -s -X GET "$API_URL/orders" \
  -H "Authorization: Bearer $TOKEN")
echo "Response: $ORDERS_RES"
echo ""

# 6. Refresh Token
echo "6️⃣  Refreshing Token..."
REFRESH_RES=$(curl -s -X POST "$API_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refreshToken\":\"$REFRESH_TOKEN\"}")

NEW_TOKEN=$(echo $REFRESH_RES | grep -o '"accessToken":"[^"]*' | cut -d'"' -f4)

if [ -z "$NEW_TOKEN" ]; then
    echo "❌ Refresh failed!"
else
    echo "✅ Token refreshed successfully!"
    echo "New Token: ${NEW_TOKEN:0:20}..."
fi
echo ""

# 7. Logout
echo "7️⃣  Logging Out..."
LOGOUT_RES=$(curl -s -X POST "$API_URL/auth/logout" \
  -H "Authorization: Bearer $NEW_TOKEN")
echo "Response: $LOGOUT_RES"
echo ""

# 8. Verify Blacklist
echo "8️⃣  Verifying Blacklist (Should fail)..."
FAIL_RES=$(curl -s -X GET "$API_URL/users/me" \
  -H "Authorization: Bearer $NEW_TOKEN")

echo "Response: $FAIL_RES"
if [[ "$FAIL_RES" == *"Token has been revoked"* ]] || [[ "$FAIL_RES" == *"Token invalid"* ]]; then
    echo "✅ Success: Token was rejected as expected."
else
    echo "⚠️  Warning: Token might still be valid or error message differs."
fi

echo ""
echo "🎉 Test Complete!"
