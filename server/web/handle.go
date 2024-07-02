package web

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func UpdateHandler(c *gin.Context) {
	fileName := c.PostForm("filename")
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

	filePath := fmt.Sprintf("data/%s", fileName)
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
		Port: 8001,
		Operate: 3,
		Address: "127.0.0.1",
		AliceFile: "data10k.csv",
		BobFile: "data10k.csv",
	}

	if err := c.BindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H { "error": err.Error() })
		return
	}

	stageShare := DoShare(params)
	stageVerify := DoVerify(params)

	c.JSON(http.StatusOK, gin.H {
		"share_info":	stageShare,
		"verify_info":	stageVerify,
	})
}