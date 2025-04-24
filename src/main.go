package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Struct to hold the secret and expiration
type secretData struct {
	Value     string
	ExpiresAt time.Time
}

var (
	secrets   = make(map[string]secretData)
	secretsMu sync.Mutex
	ttl       = 10 * time.Minute // Optional: expire secrets after 10 minutes if not read
)

func main() {
	r := gin.Default()

	// Upload a secret
	r.POST("/secret", func(c *gin.Context) {
		var req struct {
			Secret string `json:"secret" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Secret is required"})
			return
		}

		id := uuid.New().String()
		secretsMu.Lock()
		secrets[id] = secretData{
			Value:     req.Secret,
			ExpiresAt: time.Now().Add(ttl),
		}
		secretsMu.Unlock()

		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	// Retrieve and delete the secret
	r.GET("/secret/:id", func(c *gin.Context) {
		id := c.Param("id")

		secretsMu.Lock()
		data, found := secrets[id]
		if found {
			delete(secrets, id)
		}
		secretsMu.Unlock()

		if !found {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found or already used"})
			return
		}

		// Check if expired
		if time.Now().After(data.ExpiresAt) {
			c.JSON(http.StatusGone, gin.H{"error": "Secret has expired"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"secret": data.Value})
	})

	// Optional: background cleaner
	go cleanupExpiredSecrets()

	r.Run(":8080") // Listen on port 8080
}

// Periodically remove expired secrets
func cleanupExpiredSecrets() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		<-ticker.C
		now := time.Now()
		secretsMu.Lock()
		for id, data := range secrets {
			if now.After(data.ExpiresAt) {
				delete(secrets, id)
			}
		}
		secretsMu.Unlock()
	}
}
