package services

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func UpdateHandler(c *gin.Context) {
	calID := c.PostForm("id")
	party := c.PostForm("party")
	if party != "Alice" && party != "Bob" && party != "Result" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "party should be Alice or Bob"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = os.Stat(DataDir)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(DataDir, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	_, err = os.Stat(DataDir + "/" + calID)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(DataDir+"/"+calID, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	filePath := fmt.Sprintf("%s/%s/%s%s", DataDir, calID, party, "Data.csv")
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "file uploaded successfully",
		"filePath": filePath,
	})
}
