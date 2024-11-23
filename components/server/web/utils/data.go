package utils

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func getMantissa(value float64, precision int) float64 {
	if value == 0 {
		return 0.
	}

	absValue := math.Abs(value)
	exp := int(math.Floor(math.Log10(absValue)))

	return absValue / math.Pow(10, float64(exp-precision+1))
}

func CompareSignificantDigits(value1, value2 float64, precision int, scale float64) bool {
	sigDigits1 := getMantissa(value1, precision)
	sigDigits2 := getMantissa(value2, precision)
	return math.Abs(sigDigits1-sigDigits2) <= scale
}

func RunCommand(cmd string, args ...string) (string, string, int) {
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

func CompareResult(toCheckFileName, resultFileName, finalFileName string, scale float64) (int, error) {
	toCheckFile, err := os.Open(toCheckFileName)
	if err != nil {
		return -1, err
	}
	defer toCheckFile.Close()

	toCheckReader := csv.NewReader(toCheckFile)
	records, err := toCheckReader.ReadAll()
	if err != nil {
		return -1, err
	}

	results, err := os.ReadFile(resultFileName)
	if err != nil {
		return -1, err
	}

	resultData := strings.Split(string(results), "\n")
	txtValues := make([]float64, 0)
	for _, line := range resultData {
		if line != "" {
			value, err := strconv.ParseFloat(line, 64)
			if err != nil {
				return -1, err
			}
			txtValues = append(txtValues, value)
		}
	}

	checkResultFile, err := os.Create(finalFileName)
	if err != nil {
		return -1, err
	}
	defer checkResultFile.Close()

	writer := csv.NewWriter(checkResultFile)
	defer writer.Flush()

	err = writer.Write([]string{"number", "data"})
	if err != nil {
		return -1, err
	}

	errorNumber := 0
	for i, record := range records {
		if i == 0 {
			continue
		}

		number := record[0]
		dataValue, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return -1, err
		}

		result := CompareSignificantDigits(dataValue, txtValues[i-1], 6, scale)
		if !result {
			err = writer.Write([]string{number, fmt.Sprintf("%g", txtValues[i-1])})
			errorNumber += 1
		} else {
			err = writer.Write([]string{number, fmt.Sprintf("%t", result)})
		}
		// mistake := math.Abs(dataValue - txtValues[i - 1])
		// if mistake <= math.Abs(1e-6 * txtValues[i - 1]) { result = true }

		if err != nil {
			return -1, err
		}
	}

	return errorNumber, nil
}

func ParseOutputToJson(output string) (*gin.H, int, float64) {
	reComm := regexp.MustCompile(`Communication Cost:\s+(\d+)\s+bytes`)
	reTime := regexp.MustCompile(`Total Time:\s+([\d.]+)\s+ms`)

	commCostMatches := reComm.FindStringSubmatch(output)
	if commCostMatches == nil {
		return nil, -1, -1.
	}
	commCost := commCostMatches[1]

	timeMatches := reTime.FindStringSubmatch(output)
	if timeMatches == nil {
		return nil, -1, -1.
	}
	totalTime := timeMatches[1]

	parsedComm, _ := strconv.Atoi(commCost)
	parsedTime, _ := strconv.ParseFloat(totalTime, 64)

	// return &gin.H {
	// 	"comm_cost": commCost + " bytes",
	// 	"total_time": totalTime + " ms",
	// },

	return &gin.H{
			"comm_cost":  parsedComm,
			"total_time": parsedTime,
		},
		parsedComm,
		parsedTime
}
