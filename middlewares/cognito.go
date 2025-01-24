package middlewares

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type CognitoMiddlewareConfig struct {
	Region       string
	UserPoolID   string
	AppClientID  string // App Client ID for verifying ID token audience
	ResourceServerID string // Optional: Resource Server ID for verifying access token audience
}

func CognitoMiddleware(config CognitoMiddlewareConfig) fiber.Handler {
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", config.Region, config.UserPoolID)

	// Pre-fetch and cache JWKS
	keySet, err := jwk.Fetch(context.Background(), jwksURL)
	if err != nil {
		log.Fatalf("Failed to fetch JWKS: %v", err)
	}

	return func(c *fiber.Ctx) error {
		// Extract token from the Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Authorization header missing or invalid"})
		}

		// Note: The token should be an ID token for user authentication
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the token algorithm is valid
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Get the kid (key ID) from the token header
			kid, ok := token.Header["kid"].(string)
			if !ok {
				return nil, errors.New("token missing kid header")
			}

			// Get the corresponding JWK
			key, found := keySet.LookupKeyID(kid)
			if !found {
				return nil, fmt.Errorf("key not found for kid: %s", kid)
			}

			var pubKey interface{}
			if err := key.Raw(&pubKey); err != nil {
				return nil, fmt.Errorf("failed to create public key: %v", err)
			}
			return pubKey, nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		// Optionally, validate claims (audience, issuer, etc.)
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token claims"})
		}

		// Validate the audience (aud)
		if config.AppClientID != "" || config.ResourceServerID != "" {
			if aud, ok := claims["aud"].(string); ok {
				if aud != config.AppClientID && aud != config.ResourceServerID {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid audience"})
				}
			} else {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "aud claim missing"})
			}
		}

		// Validate the issuer (iss)
		expectedIssuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", config.Region, config.UserPoolID)
		if iss, ok := claims["iss"].(string); !ok || iss != expectedIssuer {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid issuer"})
		}

		// Add token claims to context
		c.Locals("claims", claims)

		// Continue to the next handler
		return c.Next()
	}
}
