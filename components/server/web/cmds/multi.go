package cmds

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"server/web/utils"
)

func DoShareMultiWorkers(params utils.VerifyParams, workers int) gin.H {
	var (
		output1 = make([]string, workers)
		output2 = make([]string, workers)
		error1  = make([]string, workers)
		error2  = make([]string, workers)
	)

	var (
		exitCode1 = make([]int, workers)
		exitCode2 = make([]int, workers)
	)

	wg := sync.WaitGroup{}
	for i := 0; i < workers; i++ {
		idx := i
		intPort, _ := strconv.Atoi(params.Port)
		intPort = intPort + idx
		wg.Add(2)

		go func() {
			defer wg.Done()
			output1[idx], error1[idx], exitCode1[idx] = utils.RunCommand(
				"./sharer",
				fmt.Sprintf("%s=%s", "ro", "1"),
				fmt.Sprintf("%s=%s", "ip", params.Address),
				fmt.Sprintf("%s=%s", "pt", strconv.Itoa(intPort)),
				fmt.Sprintf("%s=%s", "csv", fmt.Sprintf("%dAliceData.csv", idx)),
				fmt.Sprintf("%s=%s", "shr", fmt.Sprintf("%dShare.bin", idx)),
				fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("%s/%s/", DataDir, params.ID)),
			)
		}()

		go func() {
			defer wg.Done()
			output2[idx], error2[idx], exitCode2[idx] = utils.RunCommand(
				"./sharer",
				fmt.Sprintf("%s=%s", "ro", "2"),
				fmt.Sprintf("%s=%s", "ip", params.Address),
				fmt.Sprintf("%s=%s", "pt", strconv.Itoa(intPort)),
				fmt.Sprintf("%s=%s", "csv", fmt.Sprintf("%dBobData.csv", idx)),
				fmt.Sprintf("%s=%s", "shr", fmt.Sprintf("%dShare.bin", idx)),
				fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("%s/%s/", DataDir, params.ID)),
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

		if error1[i] != "" || error2[i] != "" {
			continue
		}

		_, comm1, time1 := utils.ParseOutputToJson(output1[i])
		_, comm2, time2 := utils.ParseOutputToJson(output2[i])

		if comm1 > 0 {
			summedComm1 += comm1
		}
		if comm2 > 0 {
			summedComm2 += comm2
		}

		if time1 > 0. {
			summedTime1 += time1
		}
		if time2 > 0. {
			summedTime2 += time2
		}
	}

	outputA := gin.H{}
	outputB := gin.H{}

	if len(errors1) == 0 {
		outputA = gin.H{
			"comm_cost":  summedComm1 / workers,
			"total_time": summedTime1 / float64(workers),
		}
	}

	if len(errors2) == 0 {
		outputB = gin.H{
			"comm_cost":  summedComm2 / workers,
			"total_time": summedTime2 / float64(workers),
		}
	}

	return gin.H{
		"output_alice":   outputA,
		"error_alice":    errors1,
		"exitcode_alice": exitCodes1,
		"output_bob":     outputB,
		"error_bob":      errors2,
		"exitcode_bob":   exitCodes2,
	}
}

func DoVerifyMultiWorkers(params utils.VerifyParams, workers int) gin.H {
	var (
		output1 = make([]string, workers)
		output2 = make([]string, workers)
		error1  = make([]string, workers)
		error2  = make([]string, workers)
	)

	var (
		exitCode1 = make([]int, workers)
		exitCode2 = make([]int, workers)
	)

	wg := sync.WaitGroup{}
	for i := 0; i < workers; i++ {
		idx := i
		intPort, _ := strconv.Atoi(params.Port)
		intPort = intPort + idx
		wg.Add(2)

		go func() {
			defer wg.Done()
			output1[idx], error1[idx], exitCode1[idx] = utils.RunCommand(
				"./verifier",
				fmt.Sprintf("%s=%s", "ro", "1"),
				fmt.Sprintf("%s=%s", "ip", params.Address),
				fmt.Sprintf("%s=%s", "pt", strconv.Itoa(intPort)),
				fmt.Sprintf("%s=%s", "op", strconv.Itoa(params.Operate)),
				fmt.Sprintf("%s=%s", "shr", fmt.Sprintf("%dShare.bin", idx)),
				fmt.Sprintf("%s=%s", "res", fmt.Sprintf("%dCalResult.txt", idx)),
				fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("%s/%s/", DataDir, params.ID)),
			)
		}()

		go func() {
			defer wg.Done()
			output2[idx], error2[idx], exitCode2[idx] = utils.RunCommand(
				"./verifier",
				fmt.Sprintf("%s=%s", "ro", "2"),
				fmt.Sprintf("%s=%s", "ip", params.Address),
				fmt.Sprintf("%s=%s", "pt", strconv.Itoa(intPort)),
				fmt.Sprintf("%s=%s", "op", strconv.Itoa(params.Operate)),
				fmt.Sprintf("%s=%s", "shr", fmt.Sprintf("%dShare.bin", idx)),
				fmt.Sprintf("%s=%s", "res", fmt.Sprintf("%dCalResult.txt", idx)),
				fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("%s/%s/", DataDir, params.ID)),
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

		if error1[i] != "" || error2[i] != "" {
			continue
		}

		_, comm1, time1 := utils.ParseOutputToJson(output1[i])
		_, comm2, time2 := utils.ParseOutputToJson(output2[i])

		if comm1 > 0 {
			summedComm1 += comm1
		}
		if comm2 > 0 {
			summedComm2 += comm2
		}

		if time1 > 0. {
			summedTime1 += time1
		}
		if time2 > 0. {
			summedTime2 += time2
		}
	}

	outputA := gin.H{}
	outputB := gin.H{}

	if len(errors1) == 0 {
		outputA = gin.H{
			"comm_cost":  summedComm1 / workers,
			"total_time": summedTime1 / float64(workers),
		}
	}

	if len(errors2) == 0 {
		outputB = gin.H{
			"comm_cost":  summedComm2 / workers,
			"total_time": summedTime2 / float64(workers),
		}
	}

	return gin.H{
		"output_alice":   outputA,
		"error_alice":    errors1,
		"exitcode_alice": exitCodes1,
		"output_bob":     outputB,
		"error_bob":      errors2,
		"exitcode_bob":   exitCodes2,
	}
}
