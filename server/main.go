package main

import (
	"server/web"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/verify", web.VerifyHandler)
	r.POST("/update", web.UpdateHandler)
	r.GET("/result", web.DownloadHandler)
	r.GET("/delete", web.DeleteHandler)
	
	r.Run(":8080") 
}
