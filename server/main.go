package main

import (
	"os"

	"server/web"
	"github.com/gin-gonic/gin"
)

func main() {
	port := "8080"
	if len(os.Args) > 1 { port = os.Args[1] }

	r := gin.Default()
	r.GET("/verify", web.VerifyHandler)
	r.POST("/update", web.UpdateHandler)
	r.GET("/result", web.DownloadHandler)
	r.GET("/delete", web.DeleteHandler)
	
	r.Run(":" + port) 
}
