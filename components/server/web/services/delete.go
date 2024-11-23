package services

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func DeleteHandler(c *gin.Context) {
	calID := c.Query("id")

	err := os.RemoveAll(DataDir + "/" + calID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "failed to delete"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "deleted successfully",
		"dirpath": DataDir + "/" + calID,
	})
}
