# Distributed-cloud-storage-system
Distributed cloud storage system

## 断点续传&分块上传
- 分块上传：文件切成多块，独立传输，上传完成后合并
- 断点续传：传输暂停或者异常中断后，可以基于原来进度重传

说明

- 小文件不建议分块上传
- 可以并行长传分块，并且可以无序传输
- 分块上传能够极大的提高传输效率
- 减少传输失败后重试的流量和时间

增加接口
```golang

// 初始化分块信息
http.HandleFunc("/file/mpupload/init",handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
// 上传分块
http.HandleFunc("/file/mpupload/uppart",handler.HTTPInterceptor(handler.UploadPartHandler))
// 通知分块上传完成
http.HandleFunc("/file/mpupload/complete",handler.HTTPInterceptor(handler.CompleteUploadHandler))
// 取消上传分块 未实现
http.HandleFunc("/file/mpupload/cancel",handler.HTTPInterceptor(handler.CancelUploadPartHandler))
// 查看分块上传的整体状态 未实现
http.HandleFunc("/file/mpupload/status",handler.HTTPInterceptor(handler.MultiPartUploadStatusHandler))
```
运行方式
1. 在data目录下存放需要上传的文件，修改test/test_mpupload.go里面file相关的信息，比如文件姓名，文件hash等
2. 启动程序 `go run main.go`, 另开一个terminal `go run test/test_mpupload.go`, 查看运行信息
3. 在data目录查看❤新生成的文件，应该会有一个admin+hash值的文件夹
4. 验证文件是否正确上传
   ```bash
    ls | sort -n # 进到文件夹底下，开两个进程，一个main.go一个test，然后配置文件路径，上传文件
    # 命令输出结果 1 2 3 4 排序好的文件 ps：可以不进行这一步
    cat `ls | sort -n` > a # ⭐️将所有的文件放在一个文件里⭐️
    # 计算比较hash值
    # example
    # > cat `ls | sort -n` > a
    # > sha1sum a
    # a6e3ab2cec6a2b1178a22712dc2dc548c16d1484  a
    # > sha1sum ../raftkv-master.zip
    # a6e3ab2cec6a2b1178a22712dc2dc548c16d1484  ../raftkv-master.zip
    ```
   