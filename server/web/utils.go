package web

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type VerifyParams struct {
	ID			int		`form:"id"`
	Port       	int    	`form:"port"`
	Operate    	int    	`form:"operate"`
	Address    	string 	`form:"address"`
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
	return gin.H {
		"output_alice":   	ParseOutputToJson(output1),
		"error_alice":    	error1,
		"exitcode_alice": 	exitCode1,
		"output_bob":   	ParseOutputToJson(output2),
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
	return gin.H {
		"output_alice":   	ParseOutputToJson(output1),
		"error_alice":    	error1,
		"exitcode_alice": 	exitCode1,
		"output_bob":   	ParseOutputToJson(output2),
		"error_bob":    	error2,
		"exitcode_bob": 	exitCode2,
	}
}

func CompareResult(toCheckFileName, resultFileName, finalFileName string) (int, error) {
	toCheckFile, err := os.Open(toCheckFileName)
	if err != nil { return -1, err }
	defer toCheckFile.Close()

	toCheckReader := csv.NewReader(toCheckFile)
	records, err := toCheckReader.ReadAll()
    if err != nil { return -1, err }

	results, err := os.ReadFile(resultFileName)
	if err != nil { return -1, err }

	resultData := strings.Split(string(results), "\n")
    txtValues := make([]float64, 0)
    for _, line := range resultData {
        if line != "" {
            value, err := strconv.ParseFloat(line, 64)
            if err != nil { return -1, err }
            txtValues = append(txtValues, value)
        }
    }

	checkResultFile, err := os.Create(finalFileName)
	if err != nil { return -1, err }
	defer checkResultFile.Close()

	writer := csv.NewWriter(checkResultFile)
    defer writer.Flush()

	err = writer.Write([]string { "number", "data" })
	if err != nil { return -1, err }

	errorNumber := 0
	for i, record := range records {
        if i == 0 { continue }

        number := record[0]
        dataValue, err := strconv.ParseFloat(record[1], 64)
        if err != nil { return -1, err }

		result := CompareSignificantDigits(dataValue, txtValues[i - 1], 6)
		if !result { 
			err = writer.Write([]string { number, fmt.Sprintf("%g", txtValues[i - 1]) })
			errorNumber += 1 
		} else {
			err = writer.Write([]string { number, fmt.Sprintf("%t", result) })
		}
		// mistake := math.Abs(dataValue - txtValues[i - 1])
		// if mistake <= math.Abs(1e-6 * txtValues[i - 1]) { result = true }

		if err != nil { return -1, err }
    }

	return errorNumber, nil
}

func TransferData(filePath string) error {
	csvFile, err := os.Open(filePath)
    if err != nil { return err }

    reader := csv.NewReader(csvFile)
    records, err := reader.ReadAll()
    if err != nil { return err }

	csvFile.Close()
    for i, record := range records {
        if i == 0 { continue }

        dataValue, err := strconv.ParseFloat(record[1], 64)
        if err != nil { return err }

        records[i][1] = fmt.Sprintf("%g", math.Log(dataValue))
    }

    csvFile, err = os.Create(filePath)
    if err != nil { return err }
    defer csvFile.Close()

    writer := csv.NewWriter(csvFile)
    defer writer.Flush()

	err = writer.WriteAll(records)
    if err != nil { return err }
    return nil
}

func TxtToCsv(filePath, finalFileName string) error {
    txtFile, err := os.Open(filePath)
    if err != nil { return err }
    defer txtFile.Close()

    csvFile, err := os.Create(finalFileName)
    if err != nil { return err }
    defer csvFile.Close()

    writer := bufio.NewWriter(csvFile)
    _, err = writer.WriteString("number,data\n")
    if err != nil { return err }

    scanner := bufio.NewScanner(txtFile)
    lineNumber := 1

    for scanner.Scan() {
        dataValue := scanner.Text()
        _, err = writer.WriteString(fmt.Sprintf("%d,%s\n", lineNumber, dataValue))
        if err != nil { return err }
        lineNumber++
    }

    if err := scanner.Err(); err != nil { return err }
    writer.Flush()
	return nil
}

func getMantissa(value float64, precision int) float64 {
	if value == 0 { return 0. }

	absValue := math.Abs(value)
	exp := int(math.Floor(math.Log10(absValue)))

	return absValue / math.Pow(10, float64(exp - precision + 1))
}

func CompareSignificantDigits(value1, value2 float64, precision int) bool {
	sigDigits1 := getMantissa(value1, precision)
	sigDigits2 := getMantissa(value2, precision)
	return math.Abs(sigDigits1 - sigDigits2) <= 1.
}

func ParseOutputToJson(output string) *gin.H {
	reComm := regexp.MustCompile(`Communication Cost:\s+(\d+)\s+bytes`)
	reTime := regexp.MustCompile(`Total Time:\s+([\d.]+)\s+ms`)

	commCostMatches := reComm.FindStringSubmatch(output)
    if commCostMatches == nil { return nil }
    commCost := commCostMatches[1]

	timeMatches := reTime.FindStringSubmatch(output)
    if timeMatches == nil { return nil }
    totalTime := timeMatches[1]

	return &gin.H {
		"comm_cost": commCost + " bytes",
		"total_time": totalTime + " ms",
	}
}
