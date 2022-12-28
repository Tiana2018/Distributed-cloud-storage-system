package main

import (
	"Distributed-cloud-storage-system/handler"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	// static configure
	pwd,_ := os.Getwd()
	fmt.Println("test"+pwd + " " + os.Args[0])
	http.Handle("/static/", http.FileServer(http.Dir(filepath.Join(pwd, "./"))))
	//http.Handle("/static/",
	//	http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	// 文件相关接口
	http.HandleFunc("/file/upload", handler.HTTPInterceptor(handler.UploadHandler))
	http.HandleFunc("/file/upload/suc", handler.HTTPInterceptor(handler.UploadSucHandler))
	http.HandleFunc("/file/meta", handler.HTTPInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/query", handler.HTTPInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/file/download", handler.HTTPInterceptor(handler.DownloadHandler))
	http.HandleFunc("/file/update", handler.HTTPInterceptor(handler.FileMetaUpdateHandler))
	http.HandleFunc("/file/delete", handler.HTTPInterceptor(handler.FileDeleteHandler))
	// 秒传接口
	http.HandleFunc("/file/fastupload",
		handler.HTTPInterceptor(handler.TryFastUploadHandler))
	// 下载路径
	http.HandleFunc("/file/downloadurl", handler.HTTPInterceptor(
		handler.DownloadURLHandler))

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
	// TODO 取消上传分块
	http.HandleFunc("/file/mpupload/cancel",
		handler.HTTPInterceptor(handler.CancelUploadPartHandler))
	// TODO 查看分块上传的整体状态
	http.HandleFunc("/file/mpupload/status",
		handler.HTTPInterceptor(handler.MultiPartUploadStatusHandler))

	//用户相关接口
	http.HandleFunc("/user/signup", handler.SignupHandler)
	http.HandleFunc("/user/signin", handler.SignInHandler)
	http.HandleFunc("/user/info", handler.HTTPInterceptor(handler.UserInfoHandler))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("Failed to start server,err:%s", err.Error())
	}
}
