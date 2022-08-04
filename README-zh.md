
# 火山引擎对象存储服务 Golang SDK

## 安装 SDK

SDK 支持 Go 1.13+ 版本. 运行 `go version`查看你的 Go 版本。
* 使用`go get`安装 Go SDK
  ```shell
    go get -u github.com/volcengine/ve-tos-golang-sdk/v2
  ```

## 从 v0.x/1.x 升级到 v2
更新 import path 以升级到 v2：
* 更新前
  ```go 
  import "github.com/volcengine/ve-tos-golang-sdk/tos"
  ```
* 更新后
  ```go 
  import "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
  ```
你也可以使用一些像 [mod](https://github.com/marwan-at-work/mod) 这样的开源工具来完成这件事。 以下代码展示如何使用 mod 快速升级到v2：
  ```shell
  go install github.com/marwan-at-work/mod/cmd/mod@latest
  cd your/project/dir
  mod upgrade --mod-name=github.com/volcengine/ve-tos-golang-sdk
  ```

## 使用 TOS Go SDK
* Import
  ```go 
  import "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
  ```
* 创建一个 TosClient
  ```go 
   var (
      accessKey  = "your access key"
      secretKey  = "your secret key"
      endpoint   = "your endpoint"
      region     = "your region"
   )
   client, err := tos.NewClientV2(endpoint, tos.WithRegion(region),
   tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
  ```
* 更多例子，请查看 example/ 文件夹