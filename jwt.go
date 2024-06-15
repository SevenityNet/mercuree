package main

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func isPublisherAllowed(token string) (bool, error) {
	if publisherSecret == "" {
		return true, nil
	}

	if token == "" {
		return false, nil
	}

	claims, err := validateToken(token, publisherSecret)
	if err != nil {
		return false, err
	}

	return claims != nil, nil
}

func isSubscriberAllowed(token string) ([]string, error) {
	if subscriberSecret == "" {
		return []string{}, nil
	}

	if token == "" {
		return nil, nil
	}

	claims, err := validateToken(token, subscriberSecret)
	if err != nil {
		return nil, err
	}

	if claims != nil {
		if topics, ok := claims["topics"].([]interface{}); ok {
			result := make([]string, len(topics))
			for i, topic := range topics {
				result[i] = topic.(string)
			}

			return result, nil
		}
	}

	return nil, nil
}

func validateToken(jws, secret string) (jwt.MapClaims, error) {
	token, err := jwt.ParseWithClaims(jws, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, nil
}

func extractToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header != "" {
		return header[7:] // remove "Bearer "
	}

	query := c.Query("token")
	if query != "" {
		return query
	}

	if cookie, err := c.Cookie("token"); err == nil {
		return cookie
	}

	return ""
}
