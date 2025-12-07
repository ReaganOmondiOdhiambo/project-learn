package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Configurations
var (
	PORT            string
	MONGO_URI       string
	AUTH_SERVICE_URL string
)

// Models
type Product struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	Price       float64            `json:"price" bson:"price"`
	Category    string             `json:"category" bson:"category"`
	Stock       int                `json:"stock" bson:"stock"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
}

type User struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type AuthResponse struct {
	User User `json:"user"`
}

// Database
var collection *mongo.Collection

func init() {
	// Load .env if exists
	godotenv.Load()

	PORT = getEnv("PORT", "3000")
	MONGO_URI = getEnv("MONGO_URI", "mongodb://mongodb:27017")
	AUTH_SERVICE_URL = getEnv("AUTH_SERVICE_URL", "http://auth-service:4000")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func connectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MONGO_URI))
	if err != nil {
		log.Fatal(err)
	}

	// Check connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("✅ Connected to MongoDB")
	collection = client.Database("products_db").Collection("products")
}

// Middleware
func AuthMiddleware(c *fiber.Ctx) error {
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(401).JSON(fiber.Map{"message": "Token required"})
	}

	// Validate with Auth Service
	// In Go, we'd typically use a shared JWT secret for performance,
	// but calling the service ensures blacklist check
	req, err := http.NewRequest("POST", AUTH_SERVICE_URL+"/auth/validate", nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Auth check failed"})
	}
	req.Header.Set("Authorization", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return c.Status(401).JSON(fiber.Map{"message": "Invalid token"})
	}
	defer resp.Body.Close()

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Failed to parse auth response"})
	}

	// Store user in context
	c.Locals("user", authResp.User)
	return c.Next()
}

func AdminMiddleware(c *fiber.Ctx) error {
	user := c.Locals("user").(User)
	if user.Role != "admin" {
		return c.Status(403).JSON(fiber.Map{"message": "Admin access required"})
	}
	return c.Next()
}

// Handlers
func getProducts(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Error fetching products"})
	}
	defer cursor.Close(ctx)

	var products []Product
	if err = cursor.All(ctx, &products); err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Error parsing products"})
	}

	return c.JSON(products)
}

func getProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid ID"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var product Product
	err = collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"message": "Product not found"})
	}

	return c.JSON(product)
}

func createProduct(c *fiber.Ctx) error {
	product := new(Product)
	if err := c.BodyParser(product); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid body"})
	}

	product.CreatedAt = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.InsertOne(ctx, product)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Error creating product"})
	}

	product.ID = result.InsertedID.(primitive.ObjectID)
	return c.Status(201).JSON(product)
}

func updateProduct(c *fiber.Ctx) error {
	id := c.Params("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid ID"})
	}

	var updateData bson.M
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"message": "Invalid body"})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$set": updateData},
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{"message": "Error updating product"})
	}

	if result.MatchedCount == 0 {
		return c.Status(404).JSON(fiber.Map{"message": "Product not found"})
	}

	return c.JSON(fiber.Map{"message": "Product updated"})
}

func main() {
	// Connect to DB
	connectDB()

	app := fiber.New()

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())

	// Routes
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy", "service": "product-service"})
	})

	// Public routes
	app.Get("/products", getProducts)
	app.Get("/products/:id", getProduct)

	// Protected routes
	products := app.Group("/products", AuthMiddleware)
	products.Post("/", AdminMiddleware, createProduct)
	products.Put("/:id", AdminMiddleware, updateProduct)

	log.Printf("Product Service running on port %s", PORT)
	log.Fatal(app.Listen(":" + PORT))
}
