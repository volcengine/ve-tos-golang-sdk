# ChangeLog of TOS SDK for Go

## 版本号 v2.4.1 日期：2022-10-17
- 优化：对 key 的校验规则

## 版本号 v2.4.0 日期：2022-10-13
- 新增：Bucket Lifecycle 管理接口
- 新增：Bucket Policy 管理接口
- 新增：Bucket MirrorBack 管理接口
- 新增：Object Tagging 管理接口
- 新增：FetchObjectV2 & PutFetchTaskV2
- 新增：表单上传签名接口
- 新增：修改桶的存储类型接口
- 新增：查询桶的 region 信息
- 新增：ListObjectType2 接口
- 修复：DownloadFile 支持指定下载目录

## 版本号 v2.3.3 日期： 2022-10-12
- 修复： PreSignedURL 接口 expire 过期时间问题

## 版本号 v2.3.2 日期：2022-10-09
- 优化：上传支持 / 开头的对象

## 版本号 v2.3.1 日期：2022-09-28
- 修复：UploadFile 可能出现无法正确断点续传的情况

## 版本号 v2.3.0 日期：2022-09-22
- 新增：PreSignedURL 增加 AlternativeEndpoint 参数
- 新增：域名缓存机制
- 新增：支持通过代理访问服务端

## 版本号 v2.2.2 日期：2022-9-20
- 优化：默认 transport 参数
- 修复：DownloadFile 可能出现的阻塞

## 版本号 v2.2.0 日期: 2022-9-9
- 新增: 断点续传 DownloadFile 接口
- 新增：支持自定义 Log
- 新增：支持 Cors 规则接口
- 新增：关闭重定向机制
- 新增：引入请求重试
- 新增：支持设置请求重试数量
- 新增：上传/下载支持进度条


## 版本号 v2.1.7 日期：2022-8-9
- 修复：修复了 上传类接口无法传值 nil 的 io.Reader 的问题
- 修复：修复了 ListParts 无法正确读取 Parts 字段的问题

## 版本号：v2.1.6 日期：2022-8-6
- 修复：修正了 head 类请求遇到错误时的错误信息
- 修改：允许设置 value 为空的自定义 HTTP Header 

## 版本号：v2.1.5 日期：2022-8-2
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
- 新增：初始化选项 WithHTTPTransport 设置 http.Client.Transport
- 修复：读取 Meta 不使用 map
