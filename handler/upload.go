package handler

import (
	"Distributed-cloud-storage-system/common"
	"Distributed-cloud-storage-system/config"
	dblayer "Distributed-cloud-storage-system/db"
	"Distributed-cloud-storage-system/meta"
	"Distributed-cloud-storage-system/mq"
	"Distributed-cloud-storage-system/store/ceph"
	"Distributed-cloud-storage-system/store/oss"
	"Distributed-cloud-storage-system/util"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/amz.v1/s3"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// UploadHandler deal with file upload
func UploadHandler(c *gin.Context) {
	// 返回上传html页面
	data, err := ioutil.ReadFile("./static/view/upload.html")
	if err != nil {
		fmt.Println("internal server error")
		c.String(404, `网页不存在`)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, string(data))
}

// DoUploadHandler deal with file upload
func DoUploadHandler(c *gin.Context) {
	errCode := 0
	defer func() {
		if errCode < 0 {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "Upload failed",
			})
		}else {
			c.JSON(http.StatusOK, gin.H{
				"code": errCode,
				"msg":  "上传成功",
			})
		}
	}()

	// 接受文件流及存储到本地目录
	file, head, err := c.Request.FormFile("file")
	if err != nil {
		fmt.Printf("Failed to get data", err.Error())
		errCode = -1
		return
	}
	defer file.Close()

	// 把文件内容转为[]byte
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		log.Printf("Failed to get file data, err:%s\n", err.Error())
		errCode = -2
		return
	}
	// 构建文件元信息
	fileMeta := meta.FileMeta{
		FileName: head.Filename,
		FileSha1: util.Sha1(buf.Bytes()), //　计算文件sha1
		FileSize: int64(len(buf.Bytes())),
		Location: "/tmp/" + head.Filename,
		UploadAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	// 将文件写入临时存储位置
	newFile, err := os.Create(fileMeta.Location)
	if err != nil {
		fmt.Printf("Failed to create file", err.Error())
		errCode = -3
		return
	}
	defer newFile.Close()

	nByte, err := newFile.Write(buf.Bytes())
	if int64(nByte) != fileMeta.FileSize || err != nil {
		log.Printf("Failed to save data into file, writtenSize:%d, err:%s\n", nByte, err.Error())
		errCode = -4
		return
	}

	// 同步或异步将文件转移到Ceph/OSS
	newFile.Seek(0, 0) // 游标重新回到文件头部
	if config.CurrentStoreType == common.StoreCeph {
		// 文件写入Ceph存储
		data, _ := ioutil.ReadAll(newFile)
		bucket := ceph.GetCephBucket("userfile")
		cephPath := "/ceph/" + fileMeta.FileSha1
		_ = bucket.Put(cephPath, data, "octect-stream", s3.PublicRead)
		fileMeta.Location = cephPath

	} else if config.CurrentStoreType == common.StoreOSS {
		// 文件写入OSS存储
		ossPath := "oss/" + fileMeta.FileSha1
		if config.AsyncTransferEnable {
			// 使用rabiitMQ异步存储
			data := mq.TransferData{
				FileHash:      fileMeta.FileSha1,
				CurLocation:   fileMeta.Location,
				DestLocation:  ossPath,
				DestStoreType: common.StoreOSS,
			}
			pubData, _ := json.Marshal(data)
			pubSuc := mq.Publish(
				config.TransExchangeName,
				config.TransOSSRoutingKey,
				pubData,
			)
			// 如果发送消息不成功
			if !pubSuc {
				// TODO: 当前发送转移信息失败，稍后重试
			}
		} else {
			err = oss.Bucket().PutObject(ossPath, newFile)
			if err != nil {
				fmt.Printf(err.Error())
				errCode = -5
				return
			}
			fileMeta.Location = ossPath
		}
	}

	//meta.UpdateFileMeta(fileMeta)
	_ = meta.UpdateFileMetaDB(fileMeta)

	// 更新用户文件表
	username := c.Request.FormValue("username")
	suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
	if err == nil && suc {
		errCode = 0
	} else {
		errCode = -6
	}
}

// UploadSucHandler  upload finished
func UploadSucHandler(c *gin.Context) {
	c.JSON(http.StatusOK,
		gin.H{
			"code": 0,
			"msg":  "Upload Finish!",
		})
}

// GetFileMetaHandler  获取文件元信息
func GetFileMetaHandler(c *gin.Context) {
	//filehash := c.Request.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	filehash := c.Request.FormValue("filehash")
	fMeta, err := meta.GetFileMetaDB(filehash) // 从db里面获得文件信息
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -2,
				"msg":  "Upload failed!",
			})
		return
	}
	if fMeta != nil {
		data, err := json.Marshal(fMeta)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{
					"code": -2,
					"msg":  "Upload failed!",
				})
			return
		}
		c.Data(http.StatusOK, "application/json", data)
	} else {
		c.JSON(http.StatusOK,
			gin.H{
				"code": -4,
				"msg":  "No such file",
			})
	}
}

// FileQueryHandler: 查询批量文件信息
func FileQueryHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	limitCnt, _ := strconv.Atoi(c.Request.FormValue("limit"))
	userFiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -1,
				"msg":  "Query failed!",
			})
		return
	}
	// fmt.Printf("userFiles: %v\n", userFiles)
	data, err := json.Marshal(userFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -2,
				"msg":  "Query failed!",
			})
		return
	}
	c.Data(http.StatusOK, "application/json", data)
}
func DownloadHandler(c *gin.Context) {
	fsha1 := c.Request.FormValue("filehash")
	//fm := meta.GetFileMeta(fsha1) 下载失败的原因
	fm, _ := meta.GetFileMetaDB(fsha1)
	f, err := os.Open(fm.Location)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -1,
				"msg":  "download failed!",
			})
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"code": -2,
				"msg":  "download failed!",
			})
		return
	}
	c.Writer.Header().Set("Content-Type", "application/octect-stream")
	// attachment表示文件将会提示下载到本地，而不是直接在浏览器中打开
	c.FileAttachment(fm.Location, fm.FileName)
	c.Data(http.StatusOK, "application/json", data)
}

func FileMetaUpdateHandler(c *gin.Context) {
	opType := c.Request.FormValue("op")
	filesha1 := c.Request.FormValue("filehash")
	newFileName := c.Request.FormValue("filename")

	if opType != "0" || len(newFileName) < 1 {
		c.Status(http.StatusForbidden)
		return
	}

	curFileMeta := meta.GetFileMeta(filesha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, data)
}

// FileDeleteHandler delete file
func FileDeleteHandler(c *gin.Context) {
	fileSha1 := c.Request.FormValue("filehash")
	// 删除文件
	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.Location)
	// 删除文件元信息
	meta.RemoveFileMeta(fileSha1)
	// TODO: 删除表文件信息

	c.Status(http.StatusOK)
}

// TryFastUploadHandler:尝试秒传接口
func TryFastUploadHandler(c *gin.Context) {
	// 1. 解析请求参数
	username := c.Request.FormValue("username")
	filehash := c.Request.FormValue("filehash")
	filename := c.Request.FormValue("filename")
	filesize, _ := strconv.Atoi(c.Request.FormValue("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		c.Status(http.StatusInternalServerError)
		return
	}
	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}
	// 4. 上传过则将文件信息写入用户文件表，返回成功
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		c.Data(http.StatusOK, "application/json", resp.JSONBytes())
		return
	}
}

func DownloadURLHandler(c *gin.Context) {
	filehash := c.Request.FormValue("filehash")
	// 从文件表查找记录
	row, _ := dblayer.GetFileMeta(filehash)
	// 判断文件存在OSS/本地还是Ceph

	if strings.HasPrefix(row.FileAddr.String, "/tmp") {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		tmpUrl := fmt.Sprintf("http://%s/file/download?filehash=%s&username=%s&token=%s",
			c.Request.Host, filehash, username, token)
		c.Data(http.StatusOK, "application/json",[]byte(tmpUrl))
	} else if strings.HasPrefix(row.FileAddr.String, "/ceph") {
		// TODO ceph下载url
	} else if strings.HasPrefix(row.FileAddr.String, "oss/") {
		// oss 下载url
		signedURL := oss.DownloadURL(row.FileAddr.String)
		c.Data(http.StatusOK, "application/json",[]byte(signedURL))
	}
}
