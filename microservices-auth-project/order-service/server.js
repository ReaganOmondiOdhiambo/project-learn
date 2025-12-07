const express = require('express');
const mongoose = require('mongoose');
const axios = require('axios');
const cors = require('cors');
require('dotenv').config();

const app = express();
const PORT = process.env.PORT || 6000;

// Config
const MONGO_URI = process.env.MONGO_URI || 'mongodb://mongodb:27017/orders_db';
const AUTH_SERVICE_URL = process.env.AUTH_SERVICE_URL || 'http://auth-service:4000';
const PRODUCT_SERVICE_URL = process.env.PRODUCT_SERVICE_URL || 'http://product-service:3000';
const USER_SERVICE_URL = process.env.USER_SERVICE_URL || 'http://user-service:5000';

// Middleware
app.use(cors());
app.use(express.json());

// Database
mongoose.connect(MONGO_URI)
    .then(() => console.log('✅ Connected to MongoDB (Orders)'))
    .catch(err => console.error('MongoDB connection error:', err));

// Models
const OrderSchema = new mongoose.Schema({
    userId: { type: String, required: true },
    userEmail: String,
    products: [{
        productId: String,
        name: String,
        quantity: Number,
        price: Number
    }],
    totalAmount: Number,
    status: { type: String, default: 'pending' }, // pending, completed, cancelled
    createdAt: { type: Date, default: Date.now }
});

const Order = mongoose.model('Order', OrderSchema);

// Auth Middleware
async function authenticateToken(req, res, next) {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1];

    if (!token) return res.status(401).json({ message: 'Token required' });

    try {
        // Validate with Auth Service
        const response = await axios.post(`${AUTH_SERVICE_URL}/auth/validate`, {}, {
            headers: { Authorization: `Bearer ${token}` }
        });

        req.user = response.data.user;
        req.token = token; // Keep token for downstream calls
        next();
    } catch (error) {
        console.error('Auth validation failed:', error.message);
        return res.status(401).json({ message: 'Invalid token' });
    }
}

// Routes
app.get('/health', (req, res) => {
    res.json({ status: 'healthy', service: 'order-service' });
});

/**
 * GET /orders
 * Get current user's orders
 */
app.get('/orders', authenticateToken, async (req, res) => {
    try {
        const orders = await Order.find({ userId: req.user.userId }).sort({ createdAt: -1 });
        res.json(orders);
    } catch (error) {
        res.status(500).json({ message: 'Error fetching orders' });
    }
});

/**
 * POST /orders
 * Create new order
 */
app.post('/orders', authenticateToken, async (req, res) => {
    const { products } = req.body; // Array of { productId, quantity }

    if (!products || !Array.isArray(products) || products.length === 0) {
        return res.status(400).json({ message: 'Products required' });
    }

    try {
        // 1. Validate Products & Calculate Total
        // This demonstrates service-to-service communication
        let orderProducts = [];
        let totalAmount = 0;

        for (const item of products) {
            try {
                // Call Product Service
                const productResp = await axios.get(`${PRODUCT_SERVICE_URL}/products/${item.productId}`);
                const product = productResp.data;

                if (product.stock < item.quantity) {
                    return res.status(400).json({
                        message: `Insufficient stock for product: ${product.name}`
                    });
                }

                orderProducts.push({
                    productId: product.id,
                    name: product.name,
                    price: product.price,
                    quantity: item.quantity
                });

                totalAmount += product.price * item.quantity;

            } catch (err) {
                console.error(`Error fetching product ${item.productId}:`, err.message);
                return res.status(400).json({ message: `Product not found: ${item.productId}` });
            }
        }

        // 2. Create Order
        const order = new Order({
            userId: req.user.userId,
            userEmail: req.user.email,
            products: orderProducts,
            totalAmount,
            status: 'completed' // Simplified for demo
        });

        await order.save();

        res.status(201).json(order);

    } catch (error) {
        console.error('Order creation error:', error);
        res.status(500).json({ message: 'Error creating order' });
    }
});

app.listen(PORT, () => {
    console.log(`Order Service running on port ${PORT}`);
});
