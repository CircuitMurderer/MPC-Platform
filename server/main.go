package main

import (
	"server/web"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/verify", web.VerifyHandler)
	r.POST("/update", web.UpdateHandler)
	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}
