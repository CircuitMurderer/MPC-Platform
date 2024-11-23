package services

import (
	"net/http"
	"os"
	"runtime"
	"strconv"

	"github.com/gin-gonic/gin"
	"server/web/cmds"
	"server/web/utils"
)

func VerifyHandler(c *gin.Context) {
	params := utils.VerifyParams{
		ID:      "1",
		Port:    "8001",
		Address: "127.0.0.1",

		Scale:   0,
		Operate: 3,
		Workers: 1,
	}

	if err := c.BindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	basePath := DataDir + "/" + params.ID + "/"

	_, err := os.Stat(basePath + "AliceData.csv")
	if err != nil && os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no Alice's data on server"})
		return
	}

	_, err = os.Stat(basePath + "BobData.csv")
	if err != nil && os.IsNotExist(err) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no Bob's data on server"})
		return
	}

	resultDataExist := true
	_, err = os.Stat(basePath + "ResultData.csv")
	if err != nil && os.IsNotExist(err) {
		resultDataExist = false
	}

	if params.Operate == 6 {
		err := utils.TransferData(basePath + "AliceData.csv")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if resultDataExist {
			err := utils.TransferData(basePath + "ResultData.csv")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}

		params.Operate = 2
	}

	if params.Workers < 1 || params.Workers*2 > runtime.GOMAXPROCS(0) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "workers must greater than 1 and less than " + strconv.Itoa(runtime.GOMAXPROCS(0)/2),
		})
		return
	}

	var stageShare gin.H
	var stageVerify gin.H

	if params.Workers == 1 {
		stageShare = cmds.DoShare(params)
		stageVerify = cmds.DoVerify(params)
	} else {
		utils.SplitCSV("AliceData.csv", basePath, params.Workers)
		utils.SplitCSV("BobData.csv", basePath, params.Workers)

		stageShare = cmds.DoShareMultiWorkers(params, params.Workers)
		stageVerify = cmds.DoVerifyMultiWorkers(params, params.Workers)

		utils.MergeTxt("CalResult.txt", basePath, params.Workers)
	}

	errorNumber := -2
	finalFilePath := basePath + "finalResult.csv"

	scale := 1.
	if params.Operate == 1 {
		scale = 10.
	}
	if params.Scale >= 1 {
		scale = float64(params.Scale)
	}

	if resultDataExist {
		errorNumber, err = utils.CompareResult(basePath+"ResultData.csv", basePath+"CalResult.txt", finalFilePath, scale)
	} else {
		err = utils.TxtToCsv(basePath+"CalResult.txt", finalFilePath)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"share_info":     stageShare,
		"verify_info":    stageVerify,
		"checked_errors": errorNumber,
	})
}
