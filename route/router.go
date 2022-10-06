//package route
//
//import (
//	"Distributed-cloud-storage-system/handler"
//	"github.com/gin-gonic/gin"
//)
//
//func Router() *gin.Engine {
//	// gin framework, 包括Logger, Recovery
//	router := gin.Default()
//
//	// 处理静态资源
//	router.Static("/static/", "./static")
//
//	// 不需要经过验证就能访问的接口
//	router.GET("/user/signup", handler.SignupHandler)
//	router.POST("/user/signup", handler.DoSignupHandler)
//	router.GET("/user/signin", handler.SignInHandler)
//	router.POST("/user/signin", handler.DoSignInHandler)
//	router.GET("/user/exists", handler.UserExistsHandler)
//	//加入中间件，用于校验token的拦截器
//	router.Use(handler.Authorize())
//	//use之后的所有handler都会被拦截器拦截,进行token 校验
//
//	// 用户信息查询
//	router.POST("/user/info", handler.UserInfoHandler)
//
//	// 文件相关接口
//	router.POST("/file/query", handler.FileQueryHandler)
//	router.POST("/file/update", handler.FileMetaUpdateHandler)
//	router.GET("/file/upload", handler.UploadHandler)
//	router.POST("/file/upload", handler.DoUploadHandler)
//	router.GET("/file/upload/suc", handler.UploadSucHandler)
//	router.GET("/file/meta", handler.GetFileMetaHandler)
//	router.GET("/file/download", handler.DownloadHandler)
//	router.POST("/file/download", handler.DownloadHandler)
//	router.POST("/file/delete", handler.FileDeleteHandler)
//	router.POST("/file/downloadurl", handler.DownloadURLHandler)// 下载接口
//
//	// 秒传接口
//	router.POST("/file/fastupload", handler.TryFastUploadHandler)
//
//
//	// 分块上传接口
//	// 初始化分块信息
//	router.POST("/file/mpupload/init", handler.InitialMultipartUploadHandler)
//	// 上传分块
//	router.POST("/file/mpupload/uppart", handler.UploadPartHandler)
//	// 通知分块上传完成
//	router.POST("/file/mpupload/complete", handler.CompleteUploadHandler)
//
//	return router
//}

package route

import (
	"Distributed-cloud-storage-system/handler"

	"github.com/gin-gonic/gin"
)

// Router : 网关api路由
func Router() *gin.Engine {
	router := gin.Default()

	router.Static("/static/", "./static")

	// 注册
	router.GET("/user/signup", handler.SignupHandler)
	router.POST("/user/signup", handler.DoSignupHandler)
	// 登录
	router.GET("/user/signin", handler.SigninHandler)
	router.POST("/user/signin", handler.DoSigninHandler)
	// 用户查询
	router.POST("/user/info", handler.UserInfoHandler)
	//加入中间件，用于校验token的拦截器
	router.Use(handler.Authorize())
	// 用户文件查询
	router.POST("/file/query", handler.FileQueryHandler)
	// 用户文件修改(重命名)
	router.POST("/file/update", handler.FileMetaUpdateHandler)

	// 文件相关接口
	//router.GET("/file/upload", handler.UploadHandler)
	router.POST("/file/upload", handler.DoUploadHandler)
	router.GET("/file/upload/suc", handler.UploadSucHandler)
	router.GET("/file/meta", handler.GetFileMetaHandler)
	router.GET("/file/download", handler.DownloadHandler)
	router.POST("/file/download", handler.DownloadHandler)
	router.POST("/file/delete", handler.FileDeleteHandler)
	router.POST("/file/downloadurl", handler.DownloadURLHandler) // 下载接口

	// 秒传接口
	router.POST("/file/fastupload", handler.TryFastUploadHandler)

	// 分块上传接口
	// 初始化分块信息
	router.POST("/file/mpupload/init", handler.InitialMultipartUploadHandler)
	// 上传分块
	router.POST("/file/mpupload/uppart", handler.UploadPartHandler)
	// 通知分块上传完成
	router.POST("/file/mpupload/complete", handler.CompleteUploadHandler)

	return router
}
