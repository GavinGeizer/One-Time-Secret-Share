package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

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

	r.GET("/secret/:id", func(c *gin.Context) {
		id := c.Param("id")
		secretsMu.Lock()
		data, found := secrets[id]
		if found {
			delete(secrets, id)
		}
		secretsMu.Unlock()

		if !found || time.Now().After(data.ExpiresAt) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Secret not found or already used"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"secret": data.Value})
	})

	return r
}

func TestOneTimeSecretFlow(t *testing.T) {
	router := setupRouter()

	// Step 1: Post a secret
	secret := `{"secret":"top secret"}`
	req := httptest.NewRequest("POST", "/secret", bytes.NewBuffer([]byte(secret)))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, 200, resp.Code)

	var body map[string]string
	_ = json.Unmarshal(resp.Body.Bytes(), &body)
	id, ok := body["id"]
	assert.True(t, ok)

	// Step 2: Retrieve the secret
	getReq := httptest.NewRequest("GET", "/secret/"+id, nil)
	getResp := httptest.NewRecorder()
	router.ServeHTTP(getResp, getReq)

	assert.Equal(t, 200, getResp.Code)
	var getBody map[string]string
	_ = json.Unmarshal(getResp.Body.Bytes(), &getBody)
	assert.Equal(t, "top secret", getBody["secret"])

	// Step 3: Try to retrieve again
	getReqAgain := httptest.NewRequest("GET", "/secret/"+id, nil)
	getRespAgain := httptest.NewRecorder()
	router.ServeHTTP(getRespAgain, getReqAgain)

	assert.Equal(t, 404, getRespAgain.Code)
}
