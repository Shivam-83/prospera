package salary

import (
	"log"
	"net/http"

	"github.com/doniacld/prospera/app/user"
	"github.com/gin-gonic/gin"
)

// GetSalaryBenchmarkHandler retrieves salary information for a specific userId
func GetSalaryBenchmarkHandler(c *gin.Context) {
	// Extract userId from the URL parameters
	userID := c.Query("userId")

	log.Println("GET /salary/benchmark hit, userId:", userID)

	if userID == "" {
		log.Println("Error: userId is empty")
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	}

	// Lookup the userId in SalaryInfoPerUser
	salaryInfo, exists := user.SalaryInfoPerUser[userID]
	if !exists {
		log.Println("Error: User not found in memory for userId:", userID)
		log.Println("Available users in memory:", len(user.SalaryInfoPerUser))
		// If userId does not exist, return a 404 error
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found. Please fill the form again."})
		return
	}

	log.Println("User found, returning salary info for:", userID)
	// Return the salary information as JSON if found
	c.JSON(http.StatusOK, salaryInfo)
}
