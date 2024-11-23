package services

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func DownloadHandler(c *gin.Context) {
	calID := c.Query("id")

	file, err := os.Open(fmt.Sprintf("%s/%s/finalResult.csv", DataDir, calID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "there are no result"})
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File stat error"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=resultOfID"+calID+".csv")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	io.Copy(c.Writer, file)
}
