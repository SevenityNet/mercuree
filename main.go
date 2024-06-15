package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"
)

var (
	server           = newMercureeServer()
	publisherSecret  = ""
	subscriberSecret = ""
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go startup()

	<-ctx.Done()

	log.Println("Shutting down ...")
}

func startup() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.RecoveryWithWriter(os.Stdout))
	r.Use(cors())

	r.GET("", getHello)

	// pub/sub begins here

	r.POST("/publish", publish)

	r.Use(subscriberTokenMiddleware())
	r.Use(server.subscriber())

	// pub/sub ends here

	publisherSecret = os.Getenv("PUBLISHER_JWT_SECRET")
	if publisherSecret == "" {
		log.Println("WARNING: PUBLISHER_JWT_SECRET is not set; all publishers will be allowed")
	}

	subscriberSecret = os.Getenv("SUBSCRIBER_JWT_SECRET")
	if subscriberSecret == "" {
		log.Println("WARNING: SUBSCRIBER_JWT_SECRET is not set; all subscribers will be allowed")
	}

	log.Println("Server is running on port 1560")

	r.Run(":1506")
}

func getHello(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "I am alive! 👋",
	})
}

func publish(c *gin.Context) {
	token := extractToken(c)
	allowed, err := isPublisherAllowed(token)
	if err != nil {
		panic(err)
	}

	if !allowed {
		c.JSON(403, gin.H{
			"error": "unauthorized",
		})
		return
	}

	dto := publishDto{}
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{
			"error": err.Error(),
		})
		return
	}

	server.publish(dto.Topic, dto.Data)

	c.JSON(200, gin.H{
		"message": "published",
	})
}

func subscriberTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		topics, err := isSubscriberAllowed(token)
		if err != nil {
			panic(err)
		}

		c.Set("topics", topics)

		c.Next()
	}
}

func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		corsAllowedOrigin := os.Getenv("CORS_ALLOWED_ORIGINS")

		if corsAllowedOrigin == "*" || corsAllowedOrigin == "" {
			requestDomain := c.Request.Header.Get("Origin")
			c.Writer.Header().Set("Access-Control-Allow-Origin", requestDomain)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", corsAllowedOrigin)
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, PATCH, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Baggage, Accept, Sentry-Trace")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Authorization, Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
