package tos

const (
	// Version tos-go-sdk version
	Version = "v0.2.1"
)

const (
	ACLPrivate                = "private"
	ACLPublicRead             = "public-read"
	ACLPublicReadWrite        = "public-read-write"
	ACLAuthRead               = "authenticated-read"
	ACLBucketOwnerRead        = "bucket-owner-read"
	ACLBucketOwnerFullControl = "bucket-owner-full-control"
	ACLLogDeliveryWrite       = "log-delivery-write"

	PermissionRead         = "READ"
	PermissionTypeWrite    = "WRITE"
	PermissionTypeReadAcp  = "READ_ACP"
	PermissionTypeWriteAcp = "WRITE_ACP"
	PermissionFullControl  = "FULL_CONTROL"

	//LifecycleStatusEnabled  = "Enabled"
	//LifecycleStatusDisabled = "Disabled"

	ISO8601TimeFormat = "2006-01-02T15:04:05.000Z07:00"
)

const (
	HeaderUserAgent                   = "User-Agent"
	HeaderContentLength               = "Content-Length"
	HeaderContentType                 = "Content-Type"
	HeaderContentMD5                  = "Content-MD5"
	HeaderContentSha256               = "X-Tos-Content-Sha256"
	HeaderContentLanguage             = "Content-Language"
	HeaderContentEncoding             = "Content-Encoding"
	HeaderContentDisposition          = "Content-Disposition"
	HeaderLastModified                = "Last-Modified"
	HeaderCacheControl                = "Cache-Control"
	HeaderExpires                     = "Expires"
	HeaderETag                        = "ETag"
	HeaderVersionID                   = "X-Tos-Version-Id"
	HeaderDeleteMarker                = "X-Tos-Delete-Marker"
	HeaderStorageClass                = "X-Tos-Storage-Class"
	HeaderRestore                     = "X-Tos-Restore"
	HeaderTag                         = "X-Tos-Tag"
	HeaderSSECustomerAlgorithm        = "X-Tos-Server-Side-Encryption-Customer-Algorithm"
	HeaderSSECustomerKeyMD5           = "X-Tos-Server-Side-Encryption-Customer-Key-MD5"
	HeaderSSECustomerKey              = "X-Tos-Server-Side-Encryption-Customer-Key"
	HeaderIfModifiedSince             = "If-Modified-Since"
	HeaderIfUnmodifiedSince           = "If-Unmodified-Since"
	HeaderIfMatch                     = "If-Match"
	HeaderIfNoneMatch                 = "If-None-Match"
	HeaderRange                       = "Range"
	HeaderContentRange                = "Content-Range"
	HeaderRequestID                   = "X-Tos-Request-Id"
	HeaderID2                         = "X-Tos-Id-2"
	HeaderBucketRegion                = "X-Tos-Bucket-Region"
	HeaderLocation                    = "Location"
	HeaderACL                         = "X-Tos-Acl"
	HeaderGrantFullControl            = "X-Tos-Grant-Full-Control"
	HeaderGrantRead                   = "X-Tos-Grant-Read"
	HeaderGrantReadAcp                = "X-Tos-Grant-Read-Acp"
	HeaderGrantWrite                  = "X-Tos-Grant-Write"
	HeaderGrantWriteAcp               = "X-Tos-Grant-Write-Acp"
	HeaderNextAppendOffset            = "X-Tos-Next-Append-Offset"
	HeaderObjectType                  = "X-Tos-Object-Type"
	HeaderMetadataDirective           = "X-Tos-Metadata-Directive"
	HeaderCopySource                  = "X-Tos-Copy-Source"
	HeaderCopySourceIfMatch           = "X-Tos-Copy-Source-If-Match"
	HeaderCopySourceIfNoneMatch       = "X-Tos-Copy-Source-If-None-Match"
	HeaderCopySourceIfModifiedSince   = "X-Tos-Copy-Source-If-Modified-Since"
	HeaderCopySourceIfUnmodifiedSince = "X-Tos-Copy-Source-If-Unmodified-Since"
	HeaderCopySourceRange             = "X-Tos-Copy-Source-Range"
	HeaderCopySourceVersionID         = "X-Tos-Copy-Source-Version-Id"
	HeaderWebsiteRedirectLocation     = "X-Tos-Website-Redirect-Location"
	HeaderCSType                      = "X-Tos-Cs-Type"
	HeaderMetaPrefix                  = "X-Tos-Meta-"

	// MetadataDirectiveReplace replace source object metadata when calling CopyObject
	MetadataDirectiveReplace = "REPLACE"

	// MetadataDirectiveCopy copy source object metadata when calling CopyObject
	MetadataDirectiveCopy = "COPY"
)
