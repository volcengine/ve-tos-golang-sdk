# ChangeLog of TOS SDK for Go
## 版本号 v2.7.11 日期：2025-03-25
- 修复 V1 版本 ListObject/ListVersion ObjectType 空值问题

## 版本号 v2.7.10 日期：2025-03-25
- 支持深度冷归档存储类型

## 版本号 v2.7.9 日期：2025-01-27
- 支持 QOS Policy 配置
- 支持 HNS 桶相关参数

## 版本号 v2.7.8 日期：2024-11-20
- ListObject V1 接口支持 CRC64
- 修复 Restore ExpiryDate 时间错误
- 修复 UploadFile 上传失败错误处理

## 版本号 v2.7.7 日期：2024-09-22
- 优化 HeadBucket

## 版本号 v2.7.6 日期：2024-09-18
- Dns Cache 兼容 IPV6 

## 版本号 v2.7.5 日期：2024-09-09
- GetObjectAcl 新增 IsDefault 

## 版本号 v2.7.4 日期：2024-09-05
- 新增 PutSymlink/GetSymlink 接口
- 已有对象接口支持 Object Tagging

## 版本号 v2.7.3 日期：2024-08-26
- 优化 DownloadFile part 数量限制

## 版本号 v2.7.2 日期：2024-08-05
- SetObjectMeta 支持设置 ObjectExpires
- 镜像回源接口能力补充

## 版本号 v2.7.1 日期：2024-08-01
- 新增 HNS 接口
- 修复 Checkpoint 异常情况下复用问题

## 版本号 v2.7.0 日期：2024-06-25
- 新增桶加密相关接口
- 新增桶标签相关接口
- 新增归档存储类型 StorageClassArchive
- 镜像回源支持配置将源端的头域写入自定义元数据
- GetObject 支持设置文档转码参数
- 创建桶和查询桶元数据支持 Project
- 新增柔佛 Region

## 版本号 v2.6.9 日期：2024-06-11
- 优化 CompleteMultipartUpload 返回值

## 版本号 v2.6.8 日期：2024-03-26
- 增加 GetFileStatus 接口

## 版本号 v2.6.7 日期：2024-03-26
- 预签名兼容 content-sha256 header

## 版本号 v2.6.6 日期：2024-01-29
- 增加 bucket tagging 接口
- 事件通知增加 v2 接口
- 优化对 s3 域名的判断

## 版本号 v2.6.5 日期：2023-12-21
- 增加慢日志打印
- DNS 缓存增加异步保鲜
- 新增华北2和柔佛region
- SDK 重试时增加x-tos-sdk-retry-count 头域
- 报错信息中添加 RequestUrl 和 EC 

## 版本号 v2.6.4 日期：2023-11-02

- 修复 ListMultipartUploads Prefix 未生效的问题

## 版本号 v2.6.3 日期：2023-09-15

- 修复 ContentDisposition 编码问题
- GetObject 增加图片处理 SaveAs 参数

## 版本号 v2.6.2 日期：2023-09-05

- 支持 PutObject/CopyObject/CreateMultipartUpload/CompleteMultipartUpload 增加 x-tos-forbid-overwrite 禁止覆盖头域
- PutObject/CopyObject/AppendObject 支持 if-match 头域

## 版本号 v2.6.1 日期：2023-07-10

- 新增：对象上传/下载新增 KMS 加密

## 版本号 v2.6.0 日期：2023-06-05

- 新增：支持单连接限速
- 新增：StorageClass 支持智能分层类型、冷归档
- 新增：CompleteMultipartUpload 接口支持 CompleteAll
- 新增：GetObject 支持设置图片转码参数
- 新增：支持使用自定义域名，初始化参数新增 IsCustomDomain
- 新增：支持上传回调参数
- 新增：支持镜像回源参数增强
- 新增：支持重命名单个对象
- 新增：支持取回冷归档对象
- 优化：默认重试次数调整为 3
- 修复：上传 Meta 时默认会进行 URL Decode

## 版本号 v2.5.5 日期：2023-05-11

- 新增：事件通知增加 MQ 类型

## 版本号 v2.5.4 日期：2023-04-27

- 新增：增加 rename 相关接口
- 新增：list 接口增加 user meta 字段

## 版本号 v2.5.2 日期：2023-02-23

- 修复 Windows 下 DownloadFile 失败

## 版本号 v2.5.1 日期：2023-02-23

- ListObjectType2 新增 Owner 信息
- 去除预签名配置约束
- 修复重试退避数组错误

## 版本号 v2.5.0 日期：2023-01-03

- 新增：桶跨区域复制相关接口
- 新增：桶多版本相关接口
- 新增：桶配置静态网站相关接口
- 新增：桶事件通知相关接口
- 新增：自定义域名相关接口
- 新增：断点续传复制接口
- 新增：目录分享签名接口
- 新增：列举对象v2接口

## 版本号 v2.4.6 日期：2022-12-14

- 修复：transport socket time 设置错误
- 修复：GetObject 无法指定 Response 元数据信息

## 版本号 v2.4.5 日期：2022-11-11

- 修复：修复 chunk 模式下进度条 total 是 -1

## 版本号 v2.4.4 日期：2022-11-02

- 修复：部分接口 Closer 没有正确关闭

## 版本号 v2.4.3 日期：2022-10-28

- 修复：CopyObject/SetObjectMeta 无法正确消费 VersionID
- 修复：ListObjectVersionsV2 参数无法正常消费

## 版本号 v2.4.2 日期：2022-10-19

- 优化：部分接口增加重试

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
