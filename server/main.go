package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"

	"server/web"
	"github.com/gin-gonic/gin"
)

func main() {
	port := "8080"
	procs := 2

	if len(os.Args) > 1 { port = os.Args[1] }
	if len(os.Args) > 2 { procs, _ = strconv.Atoi(os.Args[2])}

	runtime.GOMAXPROCS(procs)
	fmt.Println("[Multi-procs]", runtime.GOMAXPROCS(0))

	// gin.SetMode("release")
	r := gin.Default()
	r.GET("/verify", web.Cors(), web.VerifyHandler)
	r.POST("/update", web.Cors(), web.UpdateHandler)
	r.GET("/result", web.Cors(), web.DownloadHandler)
	r.GET("/delete", web.Cors(), web.DeleteHandler)
	
	r.Run(":" + port) 
}
