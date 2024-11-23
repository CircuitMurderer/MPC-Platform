package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/gin-gonic/gin"
	"server/web/cmds"
	"server/web/services"
)

var (
	port    string
	procs   int
	workDir string
)

func init() {
	flag.StringVar(&port, "port", "9000", "Port to run the server on (short: -p)")
	flag.IntVar(&procs, "cpus", 16, "Number of processes to use (short: -c)")
	flag.StringVar(&workDir, "dir", "data", "Data directory for the server (short: -d)")

	flag.StringVar(&port, "p", "9000", "Short for --port")
	flag.IntVar(&procs, "c", 16, "Short for --cpus")
	flag.StringVar(&workDir, "d", "data", "Short for --dir")
}

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(procs)
	fmt.Println("[Multi-procs]", runtime.GOMAXPROCS(0))

	cmds.DataDir = workDir
	services.DataDir = workDir

	// gin.SetMode("release")
	r := gin.Default()
	r.Use(services.Cors())

	r.POST("/update", services.UpdateHandler)
	r.GET("/verify", services.VerifyHandler)
	r.GET("/result", services.DownloadHandler)
	r.GET("/delete", services.DeleteHandler)

	r.Run(":" + port)
}
