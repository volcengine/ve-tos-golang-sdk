# ChangeLog of TOS SDK for Go

## 版本号：v2.1.2 日期：2022-8-2
### 变更内容
- 新增：Client.SetHTTPTransport 设置 http.Client.Transport
- 修复：读取 Meta 不使用 map

## 版本号：v2.1.1 日期：2022-7-28
### 变更内容
- 新增：初始化选项 WithHTTPTransport 设置 http.Client.Transport
- 修复：读取 Meta 不使用 map

## 版本号：v2.1.0 日期：2022-7-11
### 变更内容
- 弃用：不再使用 Bucket Handle 调用API，改为使用新增的ClientV2客户端调用API
- 弃用：不再使用 WithXXX 设置调用 API 时的参数，改为填写 XXXInput 相应字段
- 修改：User-Agent 格式改为形如 ve-tos-go-sdk/v2.1.0 (linux/amd64;go1.17.0)
- 新增：Region 已支持时，忽略 Endpoint 参数
- 新增：支持配置忽略 SSL 证书校验
- 新增：错误处理区分客户端异常与服务端异常
- 新增：HTTP 请求的基础头域和自定义元数据增加对中文的特殊处理
- 新增：从文件上传对象 PutObjectFromFile 接口
- 新增：从文件上传分片对象 UploadPartFromFile 接口
- 新增：下载对象到文件 GetObjectToFile 接口
- 新增：断点续传 UploadFile 接口
