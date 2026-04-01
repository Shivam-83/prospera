package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/doniacld/prospera/app/negotiation"
	"github.com/doniacld/prospera/app/salary"
	"github.com/doniacld/prospera/app/tips"
	"github.com/doniacld/prospera/app/user"
)

func main() {

	r := gin.Default()

	user.NewSalaryInfoPerUser()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowCredentials: false,
	}))

	// Health check — lets Render ping this endpoint to prevent cold-start sleep.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Endpoint to store user salary info
	r.POST("/salary/benchmark", salary.PostSalaryBenchmarkHandler)
	r.GET("/salary/benchmark", salary.GetSalaryBenchmarkHandler)

	// Websocket endpoints
	r.GET("/ws/salary", salary.SalaryChatWebsocketHandler)
	r.GET("/ws/negotiation", negotiation.NegotiationChatWebsocketHandler)
	r.GET("/ws/tips", tips.TipsChatWebsocketHandler)

	// Dynamic port for Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Prospera backend on port %s", port)

	err := r.Run(":" + port)
	if err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
