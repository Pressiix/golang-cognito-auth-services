package services

import (
	"context"
	"encoding/base64"
	"log"

	"crypto/hmac"
	"crypto/sha256"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/gofiber/fiber/v2"

	"go-api/utils"
)

// CognitoConfig holds the AWS Cognito configuration
type CognitoConfig struct {
	Region      string
	ClientID    string
	ClientSecret string // Leave empty if not required
}

var cognitoConfig = CognitoConfig{
	Region:      utils.LoadEnv("AWS_REGION"), // Update to your AWS region
	ClientID:     utils.LoadEnv("AWS_COGNITO_CLIENT_ID"), // Replace with your Cognito App Client ID
	ClientSecret:  utils.LoadEnv("AWS_COGNITO_CLIENT_SECRET"), // Replace with your App Client Secret (if required)
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int32  `json:"expires_in"`
}


// calculateSecretHash computes the SECRET_HASH for Cognito requests
func calculateSecretHash(username, clientID, clientSecret string) string {
	message := username + clientID
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// LoginHandler handles the login requests
func LoginHandler(c *fiber.Ctx) error {
	// Parse the request body
	var loginReq LoginRequest
	if err := c.BodyParser(&loginReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload"})
	}

	// Load AWS Config
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(cognitoConfig.Region))
	if err != nil {
		log.Println("Failed to load AWS config:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to initialize AWS client"})
	}

	// Create Cognito Identity Provider client
	cognitoClient := cognitoidentityprovider.NewFromConfig(cfg)

	// Prepare AuthParameters
	authParams := map[string]string{
		"USERNAME": loginReq.Username,
		"PASSWORD": loginReq.Password,
	}
	if cognitoConfig.ClientSecret != "" {
		authParams["SECRET_HASH"] = calculateSecretHash(loginReq.Username, cognitoConfig.ClientID, cognitoConfig.ClientSecret)
	}

	// Make the InitiateAuth API call
	output, err := cognitoClient.InitiateAuth(context.TODO(), &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:      "USER_PASSWORD_AUTH",
		ClientId:      &cognitoConfig.ClientID,
		AuthParameters: authParams,
	})
	if err != nil {
		log.Println("Failed to authenticate user:", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid username or password"})
	}

	// Prepare the response
	authResult := output.AuthenticationResult
	loginRes := LoginResponse{
		AccessToken:  *authResult.AccessToken,
		IdToken:      *authResult.IdToken,
		RefreshToken: *authResult.RefreshToken,
		ExpiresIn:    authResult.ExpiresIn,
	}

	return c.JSON(loginRes)
}
