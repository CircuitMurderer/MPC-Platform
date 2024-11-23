package utils

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
)

func TransferData(filePath string) error {
	csvFile, err := os.Open(filePath)
	if err != nil {
		return err
	}

	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	csvFile.Close()
	for i, record := range records {
		if i == 0 {
			continue
		}

		dataValue, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return err
		}

		records[i][1] = fmt.Sprintf("%g", math.Log(dataValue))
	}

	csvFile, err = os.Create(filePath)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	err = writer.WriteAll(records)
	if err != nil {
		return err
	}
	return nil
}

func TxtToCsv(filePath, finalFileName string) error {
	txtFile, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer txtFile.Close()

	csvFile, err := os.Create(finalFileName)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	writer := bufio.NewWriter(csvFile)
	_, err = writer.WriteString("number,data\n")
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(txtFile)
	lineNumber := 1

	for scanner.Scan() {
		dataValue := scanner.Text()
		_, err = writer.WriteString(fmt.Sprintf("%d,%s\n", lineNumber, dataValue))
		if err != nil {
			return err
		}
		lineNumber++
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	writer.Flush()
	return nil
}

func SplitCSV(filename, base string, parts int) error {
	csvFile, err := os.Open(base + filename)
	if err != nil {
		return err
	}
	defer csvFile.Close()

	reader := csv.NewReader(csvFile)
	header, err := reader.Read()
	if err != nil {
		return err
	}

	var records [][]string
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		records = append(records, record)
	}

	totalLines := len(records)
	linesPerPart := totalLines / parts
	if totalLines%parts != 0 {
		linesPerPart += 1
	}

	for i := 0; i < parts; i++ {
		partFilename := fmt.Sprintf("%d%s", i, filename)
		partFile, err := os.Create(base + partFilename)
		if err != nil {
			return err
		}
		defer partFile.Close()

		writer := csv.NewWriter(partFile)
		defer writer.Flush()

		err = writer.Write(header)
		if err != nil {
			return err
		}

		start := i * linesPerPart
		end := start + linesPerPart
		if end > totalLines {
			end = totalLines
		}

		for j := start; j < end; j++ {
			err := writer.Write(records[j])
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func MergeTxt(filename, base string, parts int) error {
	outFile, err := os.Create(base + filename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	for i := 0; i < parts; i++ {
		fName := fmt.Sprintf("%d%s", i, filename)
		file, err := os.Open(base + fName)
		if err != nil {
			return nil
		}
		defer file.Close()

		var numbers []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			numbers = append(numbers, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return err
		}
		for _, number := range numbers {
			if _, err := writer.WriteString(number + "\n"); err != nil {
				return fmt.Errorf("error writing to output file: %v", err)
			}
		}
	}

	return nil
}
