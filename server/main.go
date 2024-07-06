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
	r.GET("/verify", web.Cors(), web.VerifyHandler)
	r.POST("/update", web.Cors(), web.UpdateHandler)
	r.GET("/result", web.Cors(), web.DownloadHandler)
	r.GET("/delete", web.Cors(), web.DeleteHandler)
	
	r.Run(":" + port) 
}
