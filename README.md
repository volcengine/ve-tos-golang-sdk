
# Volcengine Object Storage(TOS) Go SDK

## Install TOS Go SDK

TOS Go SDK supports Go 1.13+ . Run `go version` to check your version of Golang.
* Install TOS Go SDK with `go get`
  ```shell
  go get -u github.com/volcengine/ve-tos-golang-sdk
  ```
  
## Use TOS Go SDK
* Import 
  ```go 
  import "github.com/volcengine/ve-tos-golang-sdk/tos"
  ```
* Create a client
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
* More example, see example/ folder