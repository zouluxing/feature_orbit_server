package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zouluxing/feature_orbit_server/internal/middleware"
)

// Feature is a placeholder model — replace with real GORM model in Phase 6.
type Feature struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Owner string `json:"owner"`
}

// ListFeatures godoc
// GET /api/v1/features (public)
func ListFeatures(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": []Feature{
			{ID: "feature-001", Name: "Dark Mode", Owner: "product-team"},
			{ID: "feature-002", Name: "Multi-tenant", Owner: "platform-team"},
		},
	})
}

// CreateFeature godoc
// POST /api/v1/features (requires auth)
func CreateFeature(c *gin.Context) {
	claims, _ := middleware.ClaimsFrom(c)
	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 40000, "message": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"message": "created",
		"data": Feature{
			ID:    "feature-" + claims.UserUUID[:8],
			Name:  req.Name,
			Owner: claims.Email,
		},
	})
}

// UpdateFeature godoc
// PUT /api/v1/features/:id (requires auth)
func UpdateFeature(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "updated", "data": gin.H{"id": c.Param("id")}})
}

// DeleteFeature godoc
// DELETE /api/v1/features/:id (requires admin role)
func DeleteFeature(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "deleted", "data": gin.H{"id": c.Param("id")}})
}

// GetMe godoc
// GET /api/v1/me — returns the authenticated user's claims from UMS JWT
func GetMe(c *gin.Context) {
	claims, ok := middleware.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 40004, "message": "not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"user_uuid": claims.UserUUID,
			"email":     claims.Email,
			"roles":     claims.Roles,
			"scopes":    claims.Scopes,
		},
	})
}
