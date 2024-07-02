package web

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
		Operate: 3,
		Address: "127.0.0.1",
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

	stageShare := DoShare(params)
	stageVerify := DoVerify(params)

	errorNumber := -2
	finalFilePath := basePath + "finalResult.csv"

	if resultDataExist {
		errorNumber, err = CompareResult(basePath + "ResultData.csv", basePath + "CalResult.txt", finalFilePath)
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
