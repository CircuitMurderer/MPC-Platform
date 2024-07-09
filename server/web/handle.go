package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context)  {
		// method := c.Request.Method
		origin := c.Request.Header.Get("Origin")

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", "Content-Type, AccessToken, X-CSRF-Token, Authorization, Token, X-Token, X-User-Id")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, DELETE, PUT")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}

		c.Next()
	}
}

func UpdateHandler(c *gin.Context) {
	calID := c.PostForm("id")
	party := c.PostForm("party")
	if party != "Alice" && party != "Bob" && party != "Result" {
		c.JSON(http.StatusBadRequest, gin.H { "error": "party should be Alice or Bob" })
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H { "error": err.Error() })
		return
	}

	_, err = os.Stat("data")
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll("data", os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H { "error": err.Error() })
			return
		}
	} 

	_, err = os.Stat("data/" + calID)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll("data/" + calID, os.ModePerm)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H { "error": err.Error() })
			return
		}
	} 

	filePath := fmt.Sprintf("data/%s/%s%s", calID, party, "Data.csv")
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H { "error": err.Error() })
		return
	}

	c.JSON(http.StatusOK, gin.H { 
		"message": 		"file uploaded successfully", 
		"filePath": 	filePath,
	})
}

func VerifyHandler(c *gin.Context) {
	params := VerifyParams {
		ID: 0x01,
		Port: 8001,
		Address: "127.0.0.1",

		Scale: 0,
		Operate: 3,
		Workers: 1,
	}
	basePath := "data/" + strconv.Itoa(params.ID) + "/"

	if err := c.BindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H { "error": err.Error() })
		return
	}

	_, err := os.Stat(basePath + "AliceData.csv")
	if err != nil && os.IsNotExist(err) { 
		c.JSON(http.StatusBadRequest, gin.H { "error": "no Alice's data on server" })
		return
	}

	_, err = os.Stat(basePath + "BobData.csv")
	if err != nil && os.IsNotExist(err) { 
		c.JSON(http.StatusBadRequest, gin.H { "error": "no Bob's data on server" })
		return
	}

	resultDataExist := true
	_, err = os.Stat(basePath + "ResultData.csv")
	if err != nil && os.IsNotExist(err) { resultDataExist = false }

	if params.Operate == 6 {
		err := TransferData(basePath + "AliceData.csv")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H { "error": err.Error() })
			return
		}

		if resultDataExist {
			err := TransferData(basePath + "ResultData.csv")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H { "error": err.Error() })
				return
			}
		}

		params.Operate = 2
	}

	if params.Workers < 1 || params.Workers * 2 > runtime.GOMAXPROCS(0) {
		c.JSON(http.StatusBadRequest, gin.H { 
			"error": "workers must greater than 1 and less than " + strconv.Itoa(runtime.GOMAXPROCS(0) / 2),
		})
		return
	}

	var stageShare gin.H
	var stageVerify gin.H

	if params.Workers == 1 {
		stageShare = DoShare(params)
		stageVerify = DoVerify(params)
	} else {
		SplitCSV("AliceData.csv", basePath, params.Workers)
		SplitCSV("BobData.csv", basePath, params.Workers)

		stageShare = DoShareMultiWorkers(params, params.Workers)
		stageVerify = DoVerifyMultiWorkers(params, params.Workers)

		MergeTxt("CalResult.txt", basePath, params.Workers)
	}

	errorNumber := -2
	finalFilePath := basePath + "finalResult.csv"

	scale := 1.
	if params.Operate == 1 { scale = 10. }
	if params.Scale >= 1 { scale = float64(params.Scale) }

	if resultDataExist {
		errorNumber, err = CompareResult(basePath + "ResultData.csv", basePath + "CalResult.txt", finalFilePath, scale)
	} else {
		err = TxtToCsv(basePath + "CalResult.txt", finalFilePath)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H { "error": err.Error() })
		return
	}

	c.JSON(http.StatusOK, gin.H {
		"share_info":		stageShare,
		"verify_info":		stageVerify,
		"checked_errors":	errorNumber,
	})
}

func DownloadHandler(c *gin.Context) {
	calID := c.Query("id")

	file, err := os.Open(fmt.Sprintf("data/%s/finalResult.csv", calID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H { "error": "there are no result" })
		return
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H { "error": "File stat error" })
		return
	}

	c.Header("Content-Disposition", "attachment; filename=resultOfID" + calID + ".csv")
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	io.Copy(c.Writer, file)
}

func DeleteHandler(c *gin.Context) {
	calID := c.Query("id")

	err := os.RemoveAll("data/" + calID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H { "error": "failed to delete" })
		return
	}

	c.JSON(http.StatusOK, gin.H {
		"message": 	"deleted successfully", 
		"dirpath":	"data/" + calID,	
	})
}

func DoShare(params VerifyParams) gin.H {
	var output1, output2, error1, error2 string
	var exitCode1, exitCode2 int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		output1, error1, exitCode1 = runCommand(
			"./sharer", 
			fmt.Sprintf("%s=%s", "ro", "1"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port)),
			fmt.Sprintf("%s=%s", "csv", "AliceData.csv"),
			fmt.Sprintf("%s=%s", "shr", "Share.bin"),
			fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("data/%s/", strconv.Itoa(params.ID))),
		)
	}()

	go func() {
		defer wg.Done()
		output2, error2, exitCode2 = runCommand(
			"./sharer", 
			fmt.Sprintf("%s=%s", "ro", "2"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port)),
			fmt.Sprintf("%s=%s", "csv", "BobData.csv"),
			fmt.Sprintf("%s=%s", "shr", "Share.bin"),
			fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("data/%s/", strconv.Itoa(params.ID))),
		)
	}()
	
	wg.Wait()
	outputA, _, _ := ParseOutputToJson(output1)
	outputB, _, _ := ParseOutputToJson(output2)

	return gin.H {
		"output_alice":   	outputA,
		"error_alice":    	error1,
		"exitcode_alice": 	exitCode1,
		"output_bob":   	outputB,
		"error_bob":    	error2,
		"exitcode_bob": 	exitCode2,
	}
}

func DoVerify(params VerifyParams) gin.H {
	var output1, output2, error1, error2 string
	var exitCode1, exitCode2 int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		output1, error1, exitCode1 = runCommand(
			"./verifier", 
			fmt.Sprintf("%s=%s", "ro", "1"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port)),
			fmt.Sprintf("%s=%s", "op", strconv.Itoa(params.Operate)),
			fmt.Sprintf("%s=%s", "shr", "Share.bin"),
			fmt.Sprintf("%s=%s", "res", "CalResult.txt"),
			fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("data/%s/", strconv.Itoa(params.ID))),
		)
	}()

	go func() {
		defer wg.Done()
		output2, error2, exitCode2 = runCommand(
			"./verifier", 
			fmt.Sprintf("%s=%s", "ro", "2"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port)),
			fmt.Sprintf("%s=%s", "op", strconv.Itoa(params.Operate)),
			fmt.Sprintf("%s=%s", "shr", "Share.bin"),
			fmt.Sprintf("%s=%s", "res", "CalResult.txt"),
			fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("data/%s/", strconv.Itoa(params.ID))),
		)
	}()
	
	wg.Wait()
	outputA, _, _ := ParseOutputToJson(output1)
	outputB, _, _ := ParseOutputToJson(output2)

	return gin.H {
		"output_alice":   	outputA,
		"error_alice":    	error1,
		"exitcode_alice": 	exitCode1,
		"output_bob":   	outputB,
		"error_bob":    	error2,
		"exitcode_bob": 	exitCode2,
	}
}

func DoShareMultiWorkers(params VerifyParams, workers int) gin.H {
	var (
		output1 = make([]string, workers)
		output2 = make([]string, workers)
		error1 = make([]string, workers)
		error2 = make([]string, workers)
	) 

	var (
		exitCode1 = make([]int, workers)
		exitCode2 = make([]int, workers)
	)

	wg := sync.WaitGroup {}
	for i := 0; i < workers; i++ {
		idx := i
		wg.Add(2)

		go func() {
			defer wg.Done()
			output1[idx], error1[idx], exitCode1[idx] = runCommand(
				"./sharer", 
				fmt.Sprintf("%s=%s", "ro", "1"),
				fmt.Sprintf("%s=%s", "ip", params.Address),
				fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port + idx)),
				fmt.Sprintf("%s=%s", "csv", fmt.Sprintf("%dAliceData.csv", idx)),
				fmt.Sprintf("%s=%s", "shr", fmt.Sprintf("%dShare.bin", idx)),
				fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("data/%d/", params.ID)),
			)
		}()
	
		go func() {
			defer wg.Done()
			output2[idx], error2[idx], exitCode2[idx] = runCommand(
				"./sharer", 
				fmt.Sprintf("%s=%s", "ro", "2"),
				fmt.Sprintf("%s=%s", "ip", params.Address),
				fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port + idx)),
				fmt.Sprintf("%s=%s", "csv", fmt.Sprintf("%dBobData.csv", idx)),
				fmt.Sprintf("%s=%s", "shr", fmt.Sprintf("%dShare.bin", idx)),
				fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("data/%d/", params.ID)),
			)
		}()
	}

	wg.Wait()
	summedComm1 := 0
	summedComm2 := 0
	summedTime1 := 0.
	summedTime2 := 0.

	errors1 := make([]string, 0, workers)
	errors2 := make([]string, 0, workers)
	exitCodes1 := make([]int, 0, workers)
	exitCodes2 := make([]int, 0, workers)

	for i := 0; i < workers; i++ {
		if error1[i] != "" {
			errors1 = append(errors1, fmt.Sprintf("Worker %d - Alice: %s", i, error1[i]))
			exitCodes1 = append(exitCodes1, exitCode1[i])
		}

		if error2[i] != "" {
			errors2 = append(errors2, fmt.Sprintf("Worker %d - Bob: %s", i, error2[i]))
			exitCodes2 = append(exitCodes2, exitCode2[i])
		}

		if error1[i] != "" || error2[i] != "" { continue }

		_, comm1, time1 := ParseOutputToJson(output1[i])
		_, comm2, time2 := ParseOutputToJson(output2[i])

		if comm1 > 0 { summedComm1 += comm1 }
		if comm2 > 0 { summedComm2 += comm2 }

		if time1 > 0. { summedTime1 += time1 }
		if time2 > 0. { summedTime2 += time2 }
	}

	outputA := gin.H {}
	outputB := gin.H {}

	if len(errors1) == 0 {
		outputA = gin.H { 
			"comm_cost": summedComm1 / workers,
			"total_time": summedTime1 / float64(workers),
		}
	}

	if len(errors2) == 0 {
		outputB = gin.H { 
			"comm_cost": summedComm2 / workers,
			"total_time": summedTime2 / float64(workers),
		}
	}

	return gin.H {
		"output_alice":   	outputA,
		"error_alice":    	errors1,
		"exitcode_alice": 	exitCodes1,
		"output_bob":   	outputB,
		"error_bob":    	errors2,
		"exitcode_bob": 	exitCodes2,
	}
}

func DoVerifyMultiWorkers(params VerifyParams, workers int) gin.H {
	var (
		output1 = make([]string, workers)
		output2 = make([]string, workers)
		error1 = make([]string, workers)
		error2 = make([]string, workers)
	) 

	var (
		exitCode1 = make([]int, workers)
		exitCode2 = make([]int, workers)
	)

	wg := sync.WaitGroup {}
	for i := 0; i < workers; i++ {
		idx := i
		wg.Add(2)

		go func() {
			defer wg.Done()
			output1[idx], error1[idx], exitCode1[idx] = runCommand(
				"./verifier", 
				fmt.Sprintf("%s=%s", "ro", "1"),
				fmt.Sprintf("%s=%s", "ip", params.Address),
				fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port + idx)),
				fmt.Sprintf("%s=%s", "op", strconv.Itoa(params.Operate)),
				fmt.Sprintf("%s=%s", "shr", fmt.Sprintf("%dShare.bin", idx)),
				fmt.Sprintf("%s=%s", "res", fmt.Sprintf("%dCalResult.txt", idx)),
				fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("data/%s/", strconv.Itoa(params.ID))),
			)
		}()
	
		go func() {
			defer wg.Done()
			output2[idx], error2[idx], exitCode2[idx] = runCommand(
				"./verifier", 
				fmt.Sprintf("%s=%s", "ro", "2"),
				fmt.Sprintf("%s=%s", "ip", params.Address),
				fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port + idx)),
				fmt.Sprintf("%s=%s", "op", strconv.Itoa(params.Operate)),
				fmt.Sprintf("%s=%s", "shr", fmt.Sprintf("%dShare.bin", idx)),
				fmt.Sprintf("%s=%s", "res", fmt.Sprintf("%dCalResult.txt", idx)),
				fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("data/%s/", strconv.Itoa(params.ID))),
			)
		}()
	}

	wg.Wait()
	summedComm1 := 0
	summedComm2 := 0
	summedTime1 := 0.
	summedTime2 := 0.

	errors1 := make([]string, 0, workers)
	errors2 := make([]string, 0, workers)
	exitCodes1 := make([]int, 0, workers)
	exitCodes2 := make([]int, 0, workers)

	for i := 0; i < workers; i++ {
		if error1[i] != "" {
			errors1 = append(errors1, fmt.Sprintf("Worker %d - Alice: %s", i, error1[i]))
			exitCodes1 = append(exitCodes1, exitCode1[i])
		}

		if error2[i] != "" {
			errors2 = append(errors2, fmt.Sprintf("Worker %d - Bob: %s", i, error2[i]))
			exitCodes2 = append(exitCodes2, exitCode2[i])
		}

		if error1[i] != "" || error2[i] != "" { continue }

		_, comm1, time1 := ParseOutputToJson(output1[i])
		_, comm2, time2 := ParseOutputToJson(output2[i])

		if comm1 > 0 { summedComm1 += comm1 }
		if comm2 > 0 { summedComm2 += comm2 }

		if time1 > 0. { summedTime1 += time1 }
		if time2 > 0. { summedTime2 += time2 }
	}

	outputA := gin.H {}
	outputB := gin.H {}

	if len(errors1) == 0 {
		outputA = gin.H { 
			"comm_cost": summedComm1 / workers,
			"total_time": summedTime1 / float64(workers),
		}
	}

	if len(errors2) == 0 {
		outputB = gin.H { 
			"comm_cost": summedComm2 / workers,
			"total_time": summedTime2 / float64(workers),
		}
	}

	return gin.H {
		"output_alice":   	outputA,
		"error_alice":    	errors1,
		"exitcode_alice": 	exitCodes1,
		"output_bob":   	outputB,
		"error_bob":    	errors2,
		"exitcode_bob": 	exitCodes2,
	}
}
