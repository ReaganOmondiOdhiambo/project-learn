const express = require('express');
const jwt = require('jsonwebtoken');
const bcrypt = require('bcryptjs');
const cors = require('cors');
const { body, validationResult } = require('express-validator');
const {
    initRedis,
    storeRefreshToken,
    getRefreshToken,
    deleteRefreshToken,
    blacklistToken
} = require('./utils/redis');
const { authenticateToken } = require('./middleware/auth');

require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 4000;

// Middleware
app.use(cors());
app.use(express.json());

// Initialize Redis
initRedis();

// Mock User Store (In a real app, this would be a separate DB or the Auth Service's DB)
// We'll use Redis to store user credentials for this demo to keep it self-contained
const redisClient = require('./utils/redis').initRedis();

/**
 * Helper: Get user from Redis
 */
async function getUser(email) {
    const client = await redisClient;
    const userData = await client.get(`user:${email}`);
    return userData ? JSON.parse(userData) : null;
}

/**
 * Helper: Create user in Redis
 */
async function createUser(user) {
    const client = await redisClient;
    // Store by email for lookup
    await client.set(`user:${user.email}`, JSON.stringify(user));
    // Store by ID if needed, but email is unique key here
}

// Routes

/**
 * REGISTER
 * POST /auth/register
 */
app.post('/auth/register', [
    body('email').isEmail(),
    body('password').isLength({ min: 6 }),
    body('name').notEmpty()
], async (req, res) => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
        return res.status(400).json({ errors: errors.array() });
    }

    try {
        const { email, password, name, role } = req.body;

        // Check if user exists
        const existingUser = await getUser(email);
        if (existingUser) {
            return res.status(400).json({ message: 'User already exists' });
        }

        // Hash password
        const salt = await bcrypt.genSalt(10);
        const hashedPassword = await bcrypt.hash(password, salt);

        // Create user object
        const userId = require('crypto').randomUUID();
        const user = {
            userId,
            email,
            password: hashedPassword,
            name,
            role: role || 'user',
            createdAt: new Date().toISOString()
        };

        // Save to Redis (Simulating Auth DB)
        await createUser(user);

        // In a real microservices architecture, we might emit an event here
        // or call User Service to create the profile.
        // For this demo, we'll assume User Service might lazily create profile or we just skip it for now.

        res.status(201).json({
            message: 'User registered successfully',
            userId: user.userId
        });

    } catch (error) {
        console.error('Registration error:', error);
        res.status(500).json({ message: 'Server error' });
    }
});

/**
 * LOGIN
 * POST /auth/login
 */
app.post('/auth/login', [
    body('email').isEmail(),
    body('password').exists()
], async (req, res) => {
    const errors = validationResult(req);
    if (!errors.isEmpty()) {
        return res.status(400).json({ errors: errors.array() });
    }

    try {
        const { email, password } = req.body;

        // Find user
        const user = await getUser(email);
        if (!user) {
            return res.status(400).json({ message: 'Invalid credentials' });
        }

        // Verify password
        const isMatch = await bcrypt.compare(password, user.password);
        if (!isMatch) {
            return res.status(400).json({ message: 'Invalid credentials' });
        }

        // Generate Tokens
        const payload = {
            userId: user.userId,
            email: user.email,
            role: user.role
        };

        const accessToken = jwt.sign(
            payload,
            process.env.JWT_SECRET,
            { expiresIn: process.env.JWT_EXPIRES_IN || '15m' }
        );

        const refreshToken = jwt.sign(
            payload,
            process.env.JWT_SECRET,
            { expiresIn: process.env.JWT_REFRESH_EXPIRES_IN || '7d' }
        );

        // Store refresh token in Redis
        // Parse '7d' to seconds (rough approximation for demo)
        const refreshExpiresIn = 7 * 24 * 60 * 60;
        await storeRefreshToken(user.userId, refreshToken, refreshExpiresIn);

        res.json({
            accessToken,
            refreshToken,
            user: {
                id: user.userId,
                email: user.email,
                name: user.name,
                role: user.role
            }
        });

    } catch (error) {
        console.error('Login error:', error);
        res.status(500).json({ message: 'Server error' });
    }
});

/**
 * REFRESH TOKEN
 * POST /auth/refresh
 */
app.post('/auth/refresh', async (req, res) => {
    const { refreshToken } = req.body;

    if (!refreshToken) {
        return res.status(401).json({ message: 'Refresh token required' });
    }

    try {
        // Verify token signature
        const decoded = jwt.verify(refreshToken, process.env.JWT_SECRET);

        // Check if token exists in Redis whitelist
        const storedToken = await getRefreshToken(decoded.userId);
        if (!storedToken || storedToken !== refreshToken) {
            return res.status(401).json({ message: 'Invalid refresh token' });
        }

        // Generate new Access Token
        const payload = {
            userId: decoded.userId,
            email: decoded.email,
            role: decoded.role
        };

        const accessToken = jwt.sign(
            payload,
            process.env.JWT_SECRET,
            { expiresIn: process.env.JWT_EXPIRES_IN || '15m' }
        );

        res.json({ accessToken });

    } catch (error) {
        console.error('Refresh error:', error);
        res.status(401).json({ message: 'Invalid refresh token' });
    }
});

/**
 * LOGOUT
 * POST /auth/logout
 */
app.post('/auth/logout', authenticateToken, async (req, res) => {
    try {
        // Blacklist current access token
        // Decode to get expiration
        const decoded = jwt.decode(req.token);
        const expiresIn = decoded.exp - Math.floor(Date.now() / 1000);

        if (expiresIn > 0) {
            await blacklistToken(req.token, expiresIn);
        }

        // Remove refresh token
        await deleteRefreshToken(req.user.userId);

        res.json({ message: 'Logged out successfully' });

    } catch (error) {
        console.error('Logout error:', error);
        res.status(500).json({ message: 'Server error' });
    }
});

/**
 * VALIDATE TOKEN (Internal use)
 * POST /auth/validate
 */
app.post('/auth/validate', authenticateToken, (req, res) => {
    // If middleware passes, token is valid
    res.json({
        valid: true,
        user: req.user
    });
});

/**
 * HEALTH CHECK
 */
app.get('/health', (req, res) => {
    res.json({ status: 'healthy', service: 'auth-service' });
});

// Start Server
app.listen(PORT, () => {
    console.log(`Auth Service running on port ${PORT}`);
});
