package tos

import (
	"hash/crc64"
	"os"
)

const (
	// Version tos-go-sdk version
	Version = "v2.7.12"
)

const TempFileSuffix = ".temp"
const DefaultFilePerm = os.FileMode(0644)

var DefaultCrcTable = func() *crc64.Table {
	return crc64.MakeTable(crc64.ECMA)
}

const DefaultTaskBufferSize = 100
const DefaultListMaxKeys = 1000

func getS3Endpoints() []string {
	return []string{
		"tos-s3-cn-beijing.volces.com",
		"tos-s3-cn-beijing.ivolces.com",
		"tos-s3-cn-guangzhou.volces.com",
		"tos-s3-cn-guangzhou.ivolces.com",
		"tos-s3-cn-shanghai.ivolces.com",
		"tos-s3-cn-shanghai.volces.com",
		"tos-s3-cn-beijing2.volces.com",
		"tos-s3-cn-beijing2.ivolces.com",
		"tos-s3-ap-southeast-1.volces.com",
		"tos-s3-ap-southeast-1.ivolces.com"}
}

func SupportedRegion() map[string]string {
	return map[string]string{
		"cn-beijing":     "tos-cn-beijing.volces.com",
		"cn-guangzhou":   "tos-cn-guangzhou.volces.com",
		"cn-shanghai":    "tos-cn-shanghai.volces.com",
		"cn-beijing2":    "tos-cn-beijing2.volces.com",
		"ap-southeast-1": "tos-ap-southeast-1.volces.com",
	}
}

func SupportedControlRegion() map[string]string {
	return map[string]string{
		"cn-beijing":     "tos-control-cn-beijing.volces.com",
		"cn-guangzhou":   "tos-control-cn-guangzhou.volces.com",
		"cn-shanghai":    "tos-control-cn-shanghai.volces.com",
		"cn-beijing2":    "tos-control-cn-beijing2.volces.com",
		"ap-southeast-1": "tos-control-ap-southeast-1.volces.com",
	}
}

func SupportedEndpoint() map[string]string {
	supportEndpoint := make(map[string]string)
	for key, value := range SupportedRegion() {
		supportEndpoint[value] = key
	}
	return supportEndpoint
}

const (
	defaultPreSignedURLExpires = 3600
	maxPreSignedURLExpires     = 604800
)

const (
	MaxPartSize     = 5 * 1024 * 1024 * 1024
	MinPartSize     = 5 * 1024 * 1024
	DefaultPartSize = 20 * 1024 * 1024
)

const (
	// Deprecated: use enum.ACLPrivate instead
	ACLPrivate = "private"
	// Deprecated: use enum.ACLPublicRead instead
	ACLPublicRead = "public-read"
	// Deprecated: use enum.ACLPublicReadWrite instead
	ACLPublicReadWrite = "public-read-write"
	// Deprecated: use enum.ACLAuthRead instead
	ACLAuthRead = "authenticated-read"
	// Deprecated: use enum.ACLBucketOwnerRead instead
	ACLBucketOwnerRead = "bucket-owner-read"
	// Deprecated: use enum.ACLBucketOwnerFullControl instead
	ACLBucketOwnerFullControl = "bucket-owner-full-control"
	// Deprecated: use enum.ACLLogDeliveryWrite instead
	ACLLogDeliveryWrite = "log-delivery-write"

	// Deprecated: use enum.PermissionRead instead
	PermissionRead = "READ"
	// Deprecated: use enum.PermissionWrite instead
	PermissionTypeWrite = "WRITE"
	// Deprecated: use enum.PermissionReadAcp instead
	PermissionTypeReadAcp = "READ_ACP"
	// Deprecated: use enum.PermissionWriteAcp instead
	PermissionTypeWriteAcp = "WRITE_ACP"
	// Deprecated: use enum.PermissionFullControl instead
	PermissionFullControl = "FULL_CONTROL"
)

const (
	ISO8601TimeFormat = "2006-01-02T15:04:05.000Z07:00"
)

const (
	// MetadataDirectiveReplace replace source object metadata when calling CopyObject
	MetadataDirectiveReplace = "REPLACE"

	// MetadataDirectiveCopy copy source object metadata when calling CopyObject
	MetadataDirectiveCopy = "COPY"
)
const (
	QueryPartNumber = "partNumber"
	QueryRecursive  = "recursive"
)
const (
	HeaderUserAgent                    = "User-Agent"
	HeaderContentLength                = "Content-Length"
	HeaderContentType                  = "Content-Type"
	HeaderContentMD5                   = "Content-MD5"
	HeaderContentSha256                = "X-Tos-Content-Sha256"
	HeaderContentLanguage              = "Content-Language"
	HeaderContentEncoding              = "Content-Encoding"
	HeaderContentDisposition           = "Content-Disposition"
	HeaderLastModified                 = "Last-Modified"
	HeaderCacheControl                 = "Cache-Control"
	HeaderExpires                      = "Expires"
	HeaderETag                         = "ETag"
	HeaderVersionID                    = "X-Tos-Version-Id"
	HeaderDeleteMarker                 = "X-Tos-Delete-Marker"
	HeaderStorageClass                 = "X-Tos-Storage-Class"
	HeaderAzRedundancy                 = "X-Tos-Az-Redundancy"
	HeaderRestore                      = "X-Tos-Restore"
	HeaderRestoreRequestDate           = "X-Tos-Restore-Request-Date"
	HeaderRestoreExpiryDays            = "X-Tos-Restore-Expiry-Days"
	HeaderRestoreTier                  = "X-Tos-Restore-Tier"
	HeaderTag                          = "X-Tos-Tag"
	HeaderSSECustomerAlgorithm         = "X-Tos-Server-Side-Encryption-Customer-Algorithm"
	HeaderSSECustomerKeyMD5            = "X-Tos-Server-Side-Encryption-Customer-Key-MD5"
	HeaderSSECustomerKey               = "X-Tos-Server-Side-Encryption-Customer-Key"
	HeaderServerSideEncryption         = "X-Tos-Server-Side-Encryption"
	HeaderServerSideEncryptionKmsKeyID = "X-Tos-Server-Side-Encryption-Kms-Key-Id"
	HeaderCopySourceSSECAlgorithm      = "X-Tos-Server-Side-Encryption-Customer-Algorithm"
	HeaderCopySourceSSECKeyMD5         = "X-Tos-Server-Side-Encryption-Customer-Key-MD5"
	HeaderCopySourceSSECKey            = "X-Tos-Server-Side-Encryption-Customer-Key"
	HeaderIfModifiedSince              = "If-Modified-Since"
	HeaderIfUnmodifiedSince            = "If-Unmodified-Since"
	HeaderIfMatch                      = "If-Match"
	HeaderIfNoneMatch                  = "If-None-Match"
	HeaderRange                        = "Range"
	HeaderContentRange                 = "Content-Range"
	HeaderRequestID                    = "X-Tos-Request-Id"
	HeaderID2                          = "X-Tos-Id-2"
	HeaderBucketRegion                 = "X-Tos-Bucket-Region"
	HeaderLocation                     = "Location"
	HeaderACL                          = "X-Tos-Acl"
	HeaderGrantFullControl             = "X-Tos-Grant-Full-Control"
	HeaderGrantRead                    = "X-Tos-Grant-Read"
	HeaderGrantReadAcp                 = "X-Tos-Grant-Read-Acp"
	HeaderGrantWrite                   = "X-Tos-Grant-Write"
	HeaderGrantWriteAcp                = "X-Tos-Grant-Write-Acp"
	HeaderNextAppendOffset             = "X-Tos-Next-Append-Offset"
	HeaderNextModifyOffset             = "X-Tos-Next-Modify-Offset"
	HeaderObjectType                   = "X-Tos-Object-Type"
	HeaderHashCrc64ecma                = "X-Tos-Hash-Crc64ecma"
	HeaderHashCrc32C                   = "X-Tos-Hash-Crc32c"
	HeaderMetadataDirective            = "X-Tos-Metadata-Directive"
	HeaderCopySource                   = "X-Tos-Copy-Source"
	HeaderCopySourceIfMatch            = "X-Tos-Copy-Source-If-Match"
	HeaderCopySourceIfNoneMatch        = "X-Tos-Copy-Source-If-None-Match"
	HeaderCopySourceIfModifiedSince    = "X-Tos-Copy-Source-If-Modified-Since"
	HeaderCopySourceIfUnmodifiedSince  = "X-Tos-Copy-Source-If-Unmodified-Since"
	HeaderCopySourceRange              = "X-Tos-Copy-Source-Range"
	HeaderCopySourceVersionID          = "X-Tos-Copy-Source-Version-Id"
	HeaderWebsiteRedirectLocation      = "X-Tos-Website-Redirect-Location"
	HeaderCSType                       = "X-Tos-Cs-Type"
	HeaderMetaPrefix                   = "X-Tos-Meta-"
	HeaderReplicationStatus            = "X-Tos-Replication-Status"
	HeaderProjectName                  = "X-Tos-Project-Name"
	HeaderBucketType                   = "X-Tos-Bucket-Type"
	HeaderDirectory                    = "X-Tos-Directory"
	HeaderTOSEC                        = "X-Tos-Ec"
	HeaderTosTrailer                   = "X-Tos-Trailer"
	HeaderAcceptEncoding               = "Accept-Encoding"
	HeaderExpect                       = "Expect"
	HeaderSymlinkTargetSize            = "X-Tos-Symlink-Target-Size"
	HeaderSymlinkTarget                = "X-Tos-Symlink-Target"
	HeaderSymlinkTargetBucket          = "X-Tos-Symlink-Bucket"
	HeaderTosExpiration                = "X-Tos-Expiration"
	HeaderTaggingCount                 = "X-Tos-Tagging-Count"
	HeaderTrashPath                    = "X-Tos-Trash-Path"
	HeaderRawContentLength             = "X-Tos-Raw-Content-Length"
)
