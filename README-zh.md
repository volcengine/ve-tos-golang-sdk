
# 火山引擎对象存储服务Golang SDK

## 安装SDK

SDK支持Go 1.13+ 版本. 运行 `go version`查看你的Go版本。
* 使用`go get`安装go SDK
  ```shell
    go get -u github.com/volcengine/ve-tos-golang-sdk
  ```

## 使用TOS Go SDK
* Import
  ```go 
  import "github.com/volcengine/ve-tos-golang-sdk/tos"
  ```
* 创建一个TosClient
  ```go 
   var (
      accessKey = "your Access Key"
      secretKey = "your Secret Key"
      endpoint = "your endpoint"
      region = "your region"
  )
  client, err := tos.NewClient(endpoint, tos.WithRegion(region),
  tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
  ```
* 更多例子，请查看example/文件夹