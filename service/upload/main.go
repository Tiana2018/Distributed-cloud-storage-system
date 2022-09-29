package main

import (
	"Distributed-cloud-storage-system/handler"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// static configure
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	// 文件相关接口
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/query", handler.FileQueryHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)
	// 秒传接口
	http.HandleFunc("/file/fastupload",
		handler.HTTPInterceptor(handler.TryFastUploadHandler))
	// 下载接口
	http.HandleFunc("/file/downloadurl",
		handler.HTTPInterceptor(handler.DownloadURLHandler))

	// 分块上传接口
	// 初始化分块信息
	http.HandleFunc("/file/mpupload/init",
		handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	// 上传分块
	http.HandleFunc("/file/mpupload/uppart",
		handler.HTTPInterceptor(handler.UploadPartHandler))
	// 通知分块上传完成
	http.HandleFunc("/file/mpupload/complete",
		handler.HTTPInterceptor(handler.CompleteUploadHandler))
	// 取消上传分块
	http.HandleFunc("/file/mpupload/cancel",
		handler.HTTPInterceptor(handler.CancelUploadPartHandler))
	// 查看分块上传的整体状态
	http.HandleFunc("/file/mpupload/status",
		handler.HTTPInterceptor(handler.MultiPartUploadStatusHandler))

	//用户相关接口
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	err := http.ListenAndServe(":8080", nil)
	log.Println("服务器开始启动，监听[0.0.0.0:8080]中")
	if err != nil {
		fmt.Printf("Failed to start server,err:%s", err.Error())
	}
}
