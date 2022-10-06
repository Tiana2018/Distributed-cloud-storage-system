# Distributed-cloud-storage-system

fileStore-server
基于 Golang 模拟百度网盘云存储，涉及有Golang、Redis、Ceph、阿里OSS、Gin等技术栈。

## 用户功能
- 用户注册
- 用户登陆
- token校验
## 文件功能
- 文件的增删改查
- 实现文件秒传
## 断点续传&分块上传
- 分块上传：文件切成多块，独立传输，上传完成后合并
- 断点续传：传输暂停或者异常中断后，可以基于原来进度重传

## 文件存储
- 支持ceph私有云进行存储
- 支持阿里云oss公有云存储
  - 支持文件异步转移

## 优化
- 使用Gin框架
- 微服务化和docker部署（待完成）

## 系统使用：
1. 创建对应的数据库文件
2. 配置自己的个人信息
3. `go run service/upload/main.go` 启动文件上传服务
4. 如果配置了文件异步转移，docker启动rabbitmq 执行命令`go run service/transfer/main.go` 进行监听
5. http://localhost:8080/user/signup 首次启动
