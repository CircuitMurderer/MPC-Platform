package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"server/web"
)

func main() {
	port := "8080"
	procs := 2

	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	if len(os.Args) > 2 {
		procs, _ = strconv.Atoi(os.Args[2])
	}

	runtime.GOMAXPROCS(procs)
	fmt.Println("[Multi-procs]", runtime.GOMAXPROCS(0))

	// gin.SetMode("release")
	r := gin.Default()
	r.Use(web.Cors())

	r.GET("/verify", web.VerifyHandler)
	r.POST("/update", web.UpdateHandler)
	r.GET("/result", web.DownloadHandler)
	r.GET("/delete", web.DeleteHandler)

	r.Run(":" + port)
}
