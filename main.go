package main

import (
	"log"
	"strconv"

	"go-api/middlewares"
	"go-api/services"
	"go-api/types"
	"go-api/utils"

	"github.com/gofiber/fiber/v2"

	// docs are generated by Swag CLI, you have to import them.
	// replace with your own docs folder, usually "github.com/username/reponame/docs"
	_ "go-api/docs"

	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func main() {
	 // Initialize the cache before starting the server
	 err := services.InitializeCache()
	 if err != nil {
		 log.Fatalf("Failed to initialize cache: %v", err)
	 }

	// Create a new Fiber app
	app := fiber.New()

	// http middleware -> fiber.Handler
	app.Use(adaptor.HTTPMiddleware(middlewares.LogMiddleware))

	// Configure Cognito Middleware
	cognitoMiddlewareConfig := middlewares.CognitoMiddlewareConfig{
		Region:       utils.LoadEnv("AWS_REGION"),
		UserPoolID:   utils.LoadEnv("AWS_COGNITO_USER_POOL_ID"),
		AppClientID:  utils.LoadEnv("AWS_COGNITO_CLIENT_ID"),
	}

	// Protected API group
	adminApi := app.Group("/api/admin")
	adminApi.Use(middlewares.CognitoMiddleware(cognitoMiddlewareConfig))

	// Login endpoint
	app.Post("/login", services.LoginHandler)

	adminApi.Get("/profile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "success",
		})
	})

	// Public API group
	publicApi := app.Group("/api/admin")
	publicApi.Get("/users", func(c *fiber.Ctx) error {
		users, err := services.GetAllUsers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Unable to fetch users"})
		}
		return c.JSON(users)
	})
	
	publicApi.Get("/users/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil || id < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		user, err := services.GetUserByID(id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		return c.JSON(user)
	})
	
	publicApi.Post("/users", func(c *fiber.Ctx) error {
		var user types.User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}
		if err := services.CreateUser(user); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error creating user"})
		}
		return c.Status(fiber.StatusCreated).JSON(user)
	})
	
	publicApi.Put("/users/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil || id < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		var updatedUser types.User
		if err := c.BodyParser(&updatedUser); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
		}
		if err := services.UpdateUser(id, updatedUser); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error updating user"})
		}
		return c.JSON(updatedUser)
	})
	
	publicApi.Delete("/users/:id", func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Params("id"))
		if err != nil || id < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user ID"})
		}
		if err := services.DeleteUser(id); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Error deleting user"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	// Start the server
	log.Println("Server is running on http://localhost:8080")
	log.Fatal(app.Listen(":8080"))
}
