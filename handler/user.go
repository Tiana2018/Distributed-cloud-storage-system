package handler

import (
	dblayer "Distributed-cloud-storage-system/db"
	"Distributed-cloud-storage-system/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	// 用于加密的盐值（自定义）
	pwdSalt = "*#890"
)

// SignupHandler ： 响应用户注册请求
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signup.html")
}

// DoSignupHandler : 处理注册post请求
func DoSignupHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	if len(username) < 3 || len(password) < 5 {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Invalid param",
			"code": -1,
		})
		return
	}
	// 对密码进行加盐及取Sha1值加密
	encPasswd := util.Sha1([]byte(password + pwdSalt))
	suc := dblayer.UserSignup(username, encPasswd)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":     "SIGNUP SUCCESS",
			"code":    0,
			"forward": "/user/signin",
			"data":    nil,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "SIGNUP FAILED",
			"code": -2, //todo
			"data": nil,
		})
	}
}

// SigninHandler ： 响应登陆接口
func SigninHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "http://"+c.Request.Host+"/static/view/signin.html")
}

// DoSigninHandler ： 处理登陆post请求逻辑
func DoSigninHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	encPasswd := util.Sha1([]byte(password + pwdSalt))

	// 1. 校验用户名及密码
	pwdChecked := dblayer.UserSignin(username, encPasswd)
	if !pwdChecked {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "SignIn FAILED",
			"code": -1,
		})
		return
	}

	// 2. 生成访问凭证（token）
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "SignIn generate token FAILED",
			"code": -2,
			"data": nil,
		})
		return
	}
	// 3. 登陆成功后重定向到首页
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + c.Request.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	c.Data(http.StatusOK, "octet-stream", resp.JSONBytes())
}
func UserInfoHandler(c *gin.Context) {
	// 1. 解析请求参数
	username := c.Request.FormValue("username")
	//token := r.Form.Get("token")

	//// 2. 验证token是否有效
	//isValid := IsTokenValid(token)
	//if !isValid{
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}
	// 3. 查询用户信息
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"msg":  "search user info FAILED",
			"code": -1,
		})
		return
	}
	// 4. 响应并组装用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	c.Data(http.StatusOK, "application/json", resp.JSONBytes())
}

// GenToken : 生成token
func GenToken(username string) string {
	// 40位字符：md5(username + timestamp + token_salt) + timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}

// IsTokenValid : token是否有效
func IsTokenValid(token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// time := token[:8]
	// 判断time和now的时差
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}

// UserExistsHandler ： 查询用户是否存在
func UserExistsHandler(c *gin.Context) {
	// 1. 解析请求参数
	username := c.Request.FormValue("username")

	// 3. 查询用户信息
	exists, err := dblayer.UserExist(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{
				"msg": "server error",
			})
	} else {
		c.JSON(http.StatusOK,
			gin.H{
				"msg":    "ok",
				"exists": exists,
			})
	}
}
