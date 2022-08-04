
# Volcengine Object Storage(TOS) Go SDK

## Install TOS Go SDK

TOS Go SDK supports Go 1.13+ . Run `go version` to check your version of Golang.
* Install TOS Go SDK with `go get`
  ```shell
    go get -u github.com/volcengine/ve-tos-golang-sdk/v2
  ```
  
## Migrate From v0.x/v1.x To v2
To migrate from v0.x/v1.x to v2,  you need to update the import paths to include v2.
For example:
- before
  ```go 
  import "github.com/volcengine/ve-tos-golang-sdk/tos"
  ```
- after
  ```go 
  import "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
  ```
You can also do this automatically with some open source tools like [mod](https://github.com/marwan-at-work/mod).
Following code shows how to migrate to v2 automatically with it.
  ```shell
  go install github.com/marwan-at-work/mod/cmd/mod@latest
  cd your/project/dir
  mod upgrade --mod-name=github.com/volcengine/ve-tos-golang-sdk
  ```
## Use TOS Go SDK
* Import 
  ```go 
  import "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
  ```
* Create a client
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
* More example, see example/ folder