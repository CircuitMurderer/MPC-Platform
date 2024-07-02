package web

import (
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

type VerifyParams struct {
	Port       int    `form:"port"`
	Operate    int    `form:"operate"`
	Address    string `form:"address"`
	AliceFile  string `form:"alicefile"`
	BobFile    string `form:"bobfile"`
}

func runCommand(cmd string, args ...string) (string, string, int) {
	command := exec.Command(cmd, args...)
	stdout, err := command.StdoutPipe()
	if err != nil {
		return "", err.Error(), -1
	}

	stderr, err := command.StderrPipe()
	if err != nil {
		return "", err.Error(), -1
	}

	if err := command.Start(); err != nil {
		return "", err.Error(), -1
	}

	outBytes, _ := io.ReadAll(stdout)
	errBytes, _ := io.ReadAll(stderr)
	if err = command.Wait(); err != nil {
		return "", err.Error(), -1
	}

	exitCode := command.ProcessState.ExitCode()
	return string(outBytes), string(errBytes), exitCode
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
			fmt.Sprintf("%s=%s", "csv", params.AliceFile),
			fmt.Sprintf("%s=%s", "shr", params.AliceFile + ".bin"),
			fmt.Sprintf("%s=%s", "pth", "data/"),
		)
	}()

	go func() {
		defer wg.Done()
		output2, error2, exitCode2 = runCommand(
			"./sharer", 
			fmt.Sprintf("%s=%s", "ro", "2"),
			fmt.Sprintf("%s=%s", "ip", params.Address),
			fmt.Sprintf("%s=%s", "pt", strconv.Itoa(params.Port)),
			fmt.Sprintf("%s=%s", "csv", params.BobFile),
			fmt.Sprintf("%s=%s", "shr", params.BobFile + ".bin"),
			fmt.Sprintf("%s=%s", "pth", "data/"),
		)
	}()
	
	wg.Wait()
	return gin.H {
		"output_alice":   	output1,
		"error_alice":    	error1,
		"exitcode_alice": 	exitCode1,
		"output_bob":   	output2,
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
			fmt.Sprintf("%s=%s", "shr", params.AliceFile + ".bin"),
			fmt.Sprintf("%s=%s", "res", params.AliceFile + ".txt"),
			fmt.Sprintf("%s=%s", "pth", "data/"),
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
			fmt.Sprintf("%s=%s", "shr", params.BobFile + ".bin"),
			fmt.Sprintf("%s=%s", "res", params.BobFile + ".txt"),
			fmt.Sprintf("%s=%s", "pth", "data/"),
		)
	}()
	
	wg.Wait()
	return gin.H {
		"output_alice":   	output1,
		"error_alice":    	error1,
		"exitcode_alice": 	exitCode1,
		"output_bob":   	output2,
		"error_bob":    	error2,
		"exitcode_bob": 	exitCode2,
	}
}

