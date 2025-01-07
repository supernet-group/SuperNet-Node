package server

import (
	"SuperNet-Node/api"
	"SuperNet-Node/config"
	"SuperNet-Node/middleware"
	"SuperNet-Node/server/template"
	"SuperNet-Node/utils"
	dbutils "SuperNet-Node/utils/db_utils"
	logs "SuperNet-Node/utils/log_utils"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gagliardetto/solana-go"
	"github.com/gin-gonic/gin"
)

func StartServer(serverPort string) error {
	r := gin.Default()
	r.Use(middleware.Cors())
	r.SetTrustedProxies([]string{"127.0.0.1"})
	workspace := r.Group(template.WORKSPACE)
	upload := r.Group(template.UPLOAD_file)
	// r.Any("/proxy/*proxyPath", proxyHandler)
	workspace.GET("/debugToken/:signature", getDebugToken)
	workspace.GET("/getToken/:signature", getToken)
	upload.POST("/ipfs", uploadFile)

	err := r.Run("127.0.0.1:" + serverPort)
	if err != nil {
		logs.Error(fmt.Sprintf("gin run error: %v", err))
		return err
	}
	return nil
}

func proxyHandler(c *gin.Context) {
	path := c.Param("proxyPath")

	params := c.Request.URL.Query()

	body := c.Request.Body

	url := fmt.Sprintf("http://127.0.0.1:%v"+path, config.GlobalConfig.Console.WorkPort)
	logs.Normal(fmt.Sprintf("proxy url: %v", url))

	req, err := http.NewRequest("GET", url, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	req.URL.RawQuery = params.Encode()

	for name, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

func getDebugToken(c *gin.Context) {
	signature := c.Param("signature")
	logs.Normal(fmt.Sprintf("signature: %v", signature))

	db := dbutils.GetDB()

	token, err := dbutils.Get(db, []byte("token"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ok, err := UserAuthentication(db, 100, signature, "workspace/token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if ok {
		RedirectWorkspace(c, string(token))
	} else {
		deployURL := fmt.Sprintf("http://%v:%v",
			config.GlobalConfig.Console.PublicIP,
			config.GlobalConfig.Console.SuperPort)

		logs.Normal(fmt.Sprintf("Redirect to: %v", deployURL))

		c.Redirect(http.StatusFound, deployURL)
	}
}

func getToken(c *gin.Context) {
	signature := c.Param("signature")

	db := dbutils.GetDB()

	token, err := dbutils.Get(db, []byte("token"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ok, err := UserAuthentication(db, 100, signature, "workspace/token")
	// if err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	// 	return
	// }

	ok, err := UserAuthentication(db, 1000, signature, "upload/file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> UserAuthentication %v", err.Error())})
		return
	}

	if ok {
		c.JSON(http.StatusOK, gin.H{"token": string(token)})
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "verification failed"})
}

type FileList struct {
	Path     string `json:"path" binding:"required"`
	FileType string `json:"fileType" binding:"required"` // document, folder
}

type BodyUploadFile struct {
	Signature string     `json:"signature" binding:"required"`
	ModelName string     `json:"modelName" binding:"required"`
	FileList  []FileList `json:"fileList" binding:"required"`
}

type ResUploadFile struct {
	Path string `json:"path"`
	Cid  string `json:"cid"`
}

func uploadFile(c *gin.Context) {
	var uploadFile BodyUploadFile
	if err := c.ShouldBindJSON(&uploadFile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("ShouldBindJSON: %v", err.Error())})
		return
	}

	db := dbutils.GetDB()

	buyerPublicKey, err := dbutils.Get(db, []byte("buyer"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> Get buyer %v", err.Error())})
		return
	}
	publicKeyStr := string(buyerPublicKey)
	publicKey, err := solana.PublicKeyFromBase58(publicKeyStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> PublicKeyFromBase58 %v", err.Error())})
		return
	}

	orderEndTime, err := dbutils.Get(db, []byte("orderEndTime"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> Get orderEndTime %v", err.Error())})
		return
	}
	orderEndTimeStr := string(orderEndTime)
	timeout, err := time.Parse(time.RFC3339, orderEndTimeStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> Parse orderEndTime %v", err.Error())})
		return
	}

	ok, err := UserAuthentication(db, 1000, uploadFile.Signature, "upload/file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> UserAuthentication %v", err.Error())})
		return
	}

	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "verification failed"})
		return
	}

	for index, file := range uploadFile.FileList {
		file.Path = config.GlobalConfig.Console.WorkDirectory + "/ml-workspace" + utils.EnsureLeadingSlash(file.Path)
		isExists, err := utils.PathExists(file.Path)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> PathExists %v", err.Error())})
			return
		}
		if !isExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("file not exists: %v", file.Path)})
			return
		}
		uploadFile.FileList[index].Path = file.Path
	}

	resUploadFile := []ResUploadFile{}
	for _, file := range uploadFile.FileList {
		if file.FileType == "folder" {
			files, err := utils.GetAllFiles(file.Path)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> GetAllFiles %v", err.Error())})
				return
			}

			for _, fileItem := range files {
				cid, err := api.UploadFileToIPFS(config.GlobalConfig.Console.IpfsNodeUrl, fileItem.Path, time.Until(timeout))
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> UploadFileToIPFS fileItem %v", err.Error())})
					return
				}

				destination := fmt.Sprintf(
					"/supernet/model/%v/%v%v",
					publicKey,
					uploadFile.ModelName,
					utils.EnsureLeadingSlash(utils.RemovePrefix(fileItem.Path, config.GlobalConfig.Console.WorkDirectory+"/ml-workspace")))
				err = api.RmFileInIPFS(config.GlobalConfig.Console.IpfsNodeUrl, destination)
				if err != nil {
					logs.Normal(fmt.Sprintf("> RmFileInIPFS fileItem %v", err.Error()))
					return
				}
				err = api.CopyFileInIPFS(
					config.GlobalConfig.Console.IpfsNodeUrl,
					"/ipfs/"+cid,
					destination)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> CopyFileInIPFS fileItem %v", err.Error())})
					return
				}

				resUploadFile = append(
					resUploadFile,
					ResUploadFile{Path: utils.RemovePrefix(fileItem.Path, config.GlobalConfig.Console.WorkDirectory+"/ml-workspace"), Cid: cid})
			}
		} else {
			cid, err := api.UploadFileToIPFS(config.GlobalConfig.Console.IpfsNodeUrl, file.Path, time.Until(timeout))
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> UploadFileToIPFS file %v", err.Error())})
				return
			}

			destination := fmt.Sprintf(
				"/supernet/model/%v/%v%v",
				publicKey,
				uploadFile.ModelName,
				utils.EnsureLeadingSlash(utils.RemovePrefix(file.Path, config.GlobalConfig.Console.WorkDirectory+"/ml-workspace")))
			err = api.RmFileInIPFS(config.GlobalConfig.Console.IpfsNodeUrl, destination)
			if err != nil {
				logs.Normal(fmt.Sprintf("> RmFileInIPFS fileItem %v", err.Error()))
				return
			}
			err = api.CopyFileInIPFS(
				config.GlobalConfig.Console.IpfsNodeUrl,
				"/ipfs/"+cid,
				destination)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("> CopyFileInIPFS fileItem %v", err.Error())})
				return
			}

			resUploadFile = append(
				resUploadFile,
				ResUploadFile{Path: utils.RemovePrefix(file.Path, config.GlobalConfig.Console.WorkDirectory+"/ml-workspace"), Cid: cid})
		}
	}

	logs.Normal(fmt.Sprintf("resUploadFile: %v", resUploadFile))
	c.JSON(http.StatusOK, gin.H{"resUploadFile": resUploadFile})
}

func RedirectWorkspace(c *gin.Context, token string) {
	workspaceURL := fmt.Sprintf("http://%v:%v?token=%v",
		config.GlobalConfig.Console.PublicIP,
		config.GlobalConfig.Console.SuperPort,
		token)

	logs.Normal(fmt.Sprintf("Redirect to: %v", workspaceURL))

	c.Redirect(http.StatusFound, workspaceURL)
}
