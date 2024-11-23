package cmds

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"server/web/utils"
)

func DoShare(params utils.VerifyParams) gin.H {
	var output1, output2, error1, error2 string
	var exitCode1, exitCode2 int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		output1, error1, exitCode1 = utils.RunCommand(
			"./sharer",
			fmt.Sprintf("%s=%s", "ro", "1"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", params.Port),
			fmt.Sprintf("%s=%s", "csv", "AliceData.csv"),
			fmt.Sprintf("%s=%s", "shr", "Share.bin"),
			fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("%s/%s/", DataDir, params.ID)),
		)
	}()

	go func() {
		defer wg.Done()
		output2, error2, exitCode2 = utils.RunCommand(
			"./sharer",
			fmt.Sprintf("%s=%s", "ro", "2"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", params.Port),
			fmt.Sprintf("%s=%s", "csv", "BobData.csv"),
			fmt.Sprintf("%s=%s", "shr", "Share.bin"),
			fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("%s/%s/", DataDir, params.ID)),
		)
	}()

	wg.Wait()
	outputA, _, _ := utils.ParseOutputToJson(output1)
	outputB, _, _ := utils.ParseOutputToJson(output2)

	return gin.H{
		"output_alice":   outputA,
		"error_alice":    error1,
		"exitcode_alice": exitCode1,
		"output_bob":     outputB,
		"error_bob":      error2,
		"exitcode_bob":   exitCode2,
	}
}

func DoVerify(params utils.VerifyParams) gin.H {
	var output1, output2, error1, error2 string
	var exitCode1, exitCode2 int

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		output1, error1, exitCode1 = utils.RunCommand(
			"./verifier",
			fmt.Sprintf("%s=%s", "ro", "1"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", params.Port),
			fmt.Sprintf("%s=%s", "op", strconv.Itoa(params.Operate)),
			fmt.Sprintf("%s=%s", "shr", "Share.bin"),
			fmt.Sprintf("%s=%s", "res", "CalResult.txt"),
			fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("%s/%s/", DataDir, params.ID)),
		)
	}()

	go func() {
		defer wg.Done()
		output2, error2, exitCode2 = utils.RunCommand(
			"./verifier",
			fmt.Sprintf("%s=%s", "ro", "2"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", params.Port),
			fmt.Sprintf("%s=%s", "op", strconv.Itoa(params.Operate)),
			fmt.Sprintf("%s=%s", "shr", "Share.bin"),
			fmt.Sprintf("%s=%s", "res", "CalResult.txt"),
			fmt.Sprintf("%s=%s", "pth", fmt.Sprintf("%s/%s/", DataDir, params.ID)),
		)
	}()

	wg.Wait()
	outputA, _, _ := utils.ParseOutputToJson(output1)
	outputB, _, _ := utils.ParseOutputToJson(output2)

	return gin.H{
		"output_alice":   outputA,
		"error_alice":    error1,
		"exitcode_alice": exitCode1,
		"output_bob":     outputB,
		"error_bob":      error2,
		"exitcode_bob":   exitCode2,
	}
}
