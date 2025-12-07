/**
 * AUTHENTICATION MIDDLEWARE
 * =========================
 * Validates JWT tokens and checks blacklist
 */

const jwt = require('jsonwebtoken');
const { isTokenBlacklisted } = require('../utils/redis');

/**
 * Verify JWT token and attach user to request
 */
async function authenticateToken(req, res, next) {
    try {
        // Get token from Authorization header
        const authHeader = req.headers['authorization'];
        const token = authHeader && authHeader.split(' ')[1]; // Bearer TOKEN

        if (!token) {
            return res.status(401).json({
                error: 'Access denied',
                message: 'No token provided'
            });
        }

        // Check if token is blacklisted
        const blacklisted = await isTokenBlacklisted(token);
        if (blacklisted) {
            return res.status(401).json({
                error: 'Token invalid',
                message: 'Token has been revoked'
            });
        }

        // Verify token
        const decoded = jwt.verify(token, process.env.JWT_SECRET);

        // Attach user info to request
        req.user = {
            userId: decoded.userId,
            email: decoded.email,
            role: decoded.role || 'user'
        };

        // Attach token for potential blacklisting
        req.token = token;

        next();
    } catch (error) {
        if (error.name === 'TokenExpiredError') {
            return res.status(401).json({
                error: 'Token expired',
                message: 'Please refresh your token'
            });
        }

        if (error.name === 'JsonWebTokenError') {
            return res.status(401).json({
                error: 'Invalid token',
                message: 'Token is malformed or invalid'
            });
        }

        console.error('Auth middleware error:', error);
        res.status(500).json({
            error: 'Authentication failed',
            message: 'Internal server error'
        });
    }
}

/**
 * Check if user has required role
 */
function requireRole(role) {
    return (req, res, next) => {
        if (!req.user) {
            return res.status(401).json({
                error: 'Unauthorized',
                message: 'Authentication required'
            });
        }

        if (req.user.role !== role && req.user.role !== 'admin') {
            return res.status(403).json({
                error: 'Forbidden',
                message: `Requires ${role} role`
            });
        }

        next();
    };
}

module.exports = {
    authenticateToken,
    requireRole
};
