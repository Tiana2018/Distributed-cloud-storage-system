package handler

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	rPool "Distributed-cloud-storage-system/cache/redis"
	dblayer "Distributed-cloud-storage-system/db"
	"Distributed-cloud-storage-system/util"
)

type MultipartUploadinfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

// 初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "param invalid", nil).JSONBytes())
	}

	//2. 获得redis的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//3. 生成分块上传的初始化信息
	upInfo := MultipartUploadinfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()),
		ChunkSize:  500 * 1024, // 5MB --> 500KB
		ChunkCount: int(math.Ceil(float64(filesize) / (500 * 1024))),
	}

	//4. 将初始化信息写入到redis缓存
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "chunkcount", upInfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filehash", upInfo.FileHash)
	rConn.Do("HSET", "MP_"+upInfo.UploadID, "filesize", upInfo.FileSize)

	//5. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "ok", upInfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	// 2 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3 获得文件句柄，用于存储分块内容
	fpath := "data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744) //必须先创建目录，不然文件路径不存在会报错
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Upload part failed", err).JSONBytes())
		return
	}
	defer fd.Close()
	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}
	// 4 更新redis缓存状态
	rConn.Do("HEST", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	// 5 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
}

// 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	filename := r.Form.Get("filename")

	// 2. 获得redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	// 3. 通过uploadid查询redis并判断是否所有分块上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+uploadID))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Complete upload failed", nil).JSONBytes())
		return
	}
	totalCount := 0
	chunkCount := 0
	for i := 0; i < len(data); i += 2 { //通过跳表查出来的value是在一个slice里面的
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunckcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount {
		w.Write(util.NewRespMsg(-2, "invalid request", nil).JSONBytes())
		return
	}
	// 4. TODO:合并分块

	// 5. 更新唯一文件表以及用户文件表
	dblayer.OnFileUploadFinished(filehash, filename, int64(filesize), "")
	dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))

	// 6. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "ok", nil).JSONBytes())
}

// 取消上传分块
func CancelUploadPartHandler(w http.ResponseWriter, r *http.Request) {

}

// 查看分块上传的整体状态
func MultiPartUploadStatusHandler(w http.ResponseWriter, r *http.Request) {

}
