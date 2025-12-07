/**
 * REDIS CLIENT - Token Storage & Blacklist
 * =========================================
 * This module manages Redis connections for:
 * - Token blacklisting (logout)
 * - Refresh token storage
 * - Session management
 */

const redis = require('redis');

let redisClient = null;

/**
 * Initialize Redis connection
 */
async function initRedis() {
    if (redisClient) {
        return redisClient;
    }

    redisClient = redis.createClient({
        socket: {
            host: process.env.REDIS_HOST || 'localhost',
            port: process.env.REDIS_PORT || 6379
        },
        password: process.env.REDIS_PASSWORD || undefined
    });

    redisClient.on('error', (err) => {
        console.error('Redis Client Error:', err);
    });

    redisClient.on('connect', () => {
        console.log('✅ Connected to Redis');
    });

    await redisClient.connect();
    return redisClient;
}

/**
 * Add token to blacklist (for logout)
 * @param {string} token - JWT token to blacklist
 * @param {number} expiresIn - Token expiration time in seconds
 */
async function blacklistToken(token, expiresIn) {
    const client = await initRedis();
    // Store token with expiration matching JWT expiry
    await client.setEx(`blacklist:${token}`, expiresIn, 'true');
    console.log(`🚫 Token blacklisted for ${expiresIn}s`);
}

/**
 * Check if token is blacklisted
 * @param {string} token - JWT token to check
 * @returns {boolean} - True if blacklisted
 */
async function isTokenBlacklisted(token) {
    const client = await initRedis();
    const result = await client.get(`blacklist:${token}`);
    return result !== null;
}

/**
 * Store refresh token
 * @param {string} userId - User ID
 * @param {string} refreshToken - Refresh token
 * @param {number} expiresIn - Expiration time in seconds
 */
async function storeRefreshToken(userId, refreshToken, expiresIn) {
    const client = await initRedis();
    await client.setEx(`refresh:${userId}`, expiresIn, refreshToken);
    console.log(`💾 Refresh token stored for user ${userId}`);
}

/**
 * Get refresh token for user
 * @param {string} userId - User ID
 * @returns {string|null} - Refresh token or null
 */
async function getRefreshToken(userId) {
    const client = await initRedis();
    return await client.get(`refresh:${userId}`);
}

/**
 * Delete refresh token (logout)
 * @param {string} userId - User ID
 */
async function deleteRefreshToken(userId) {
    const client = await initRedis();
    await client.del(`refresh:${userId}`);
    console.log(`🗑️  Refresh token deleted for user ${userId}`);
}

/**
 * Close Redis connection
 */
async function closeRedis() {
    if (redisClient) {
        await redisClient.quit();
        redisClient = null;
        console.log('Redis connection closed');
    }
}

module.exports = {
    initRedis,
    blacklistToken,
    isTokenBlacklisted,
    storeRefreshToken,
    getRefreshToken,
    deleteRefreshToken,
    closeRedis
};
