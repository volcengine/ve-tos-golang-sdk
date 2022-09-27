package tos

import (
	"io"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

type Grantee struct {
	ID          string `json:"ID,omitempty"`
	DisplayName string `json:"DisplayName,omitempty"`
	Type        string `json:"Type,omitempty"`
	URI         string `json:"Canned,omitempty"`
}

type GranteeV2 struct {
	ID          string
	DisplayName string
	Type        enum.GranteeType
	Canned      enum.CannedType
}

type GrantV2 struct {
	GranteeV2  GranteeV2           `json:"Grantee,omitempty"`
	Permission enum.PermissionType `json:"Permission,omitempty"`
}

type Grant struct {
	Grantee    Grantee             `json:"Grantee,omitempty"`
	Permission enum.PermissionType `json:"Permission,omitempty"`
}

type ObjectAclGrant struct {
	ACL              string `json:"ACL,omitempty"`
	GrantFullControl string `json:"GrantFullControl,omitempty"`
	GrantRead        string `json:"GrantRead,omitempty"`
	GrantReadAcp     string `json:"GrantReadAcp,omitempty"`
	// Deprecated: GrantWrite will be ignored
	GrantWrite    string `json:"GrantWrite,omitempty"`
	GrantWriteAcp string `json:"GrantWriteAcp,omitempty"`
}

type ObjectAclRules struct {
	Owner  Owner   `json:"Owner,omitempty"`
	Grants []Grant `json:"Grants,omitempty"`
}

type GetObjectAclOutput struct {
	RequestInfo `json:"-"`
	VersionID   string  `json:"VersionId,omitempty"`
	Owner       Owner   `json:"Owner,omitempty"`
	Grants      []Grant `json:"Grants,omitempty"`
}

type GetObjectACLInput struct {
	Bucket    string
	Key       string
	VersionID string `location:"query" locationName:"versionId"`
}

type GetObjectACLOutput struct {
	RequestInfo `json:"-"`
	VersionID   string    `json:"VersionID,omitempty"`
	Owner       Owner     `json:"Owner,omitempty"`
	Grants      []GrantV2 `json:"Grants,omitempty"`
}

// PutObjectAclInput AclGrant, AclRules can not set both.
type PutObjectAclInput struct {
	Key       string          `json:"Key,omitempty"`       // the object, required
	VersionID string          `json:"VersionId,omitempty"` // the version id of the object, optional
	AclGrant  *ObjectAclGrant `json:"AclGrant,omitempty"`  // set acl by header
	AclRules  *ObjectAclRules `json:"AclRules,omitempty"`  // set acl by rules
}

type PutObjectACLInput struct {
	Bucket           string
	Key              string       // the object, required
	VersionID        string       `location:"query" locationName:"versionId"`                 // optional
	ACL              enum.ACLType `location:"header" locationName:"X-Tos-Acl"`                // optional
	GrantFullControl string       `location:"header" locationName:"X-Tos-Grant-Full-Control"` // optional
	GrantRead        string       `location:"header" locationName:"X-Tos-Grant-Read"`         // optional
	GrantReadAcp     string       `location:"header" locationName:"X-Tos-Grant-Read-Acp"`     // optional
	GrantWrite       string       `location:"header" locationName:"X-Tos-Grant-Write"`        // optional
	GrantWriteAcp    string       `location:"header" locationName:"X-Tos-Grant-Write-Acp"`    // optional
	Owner            Owner
	Grants           []GrantV2
}

type PutObjectAclOutput struct {
	RequestInfo `json:"-"`
}

type PutObjectACLOutput struct {
	PutObjectAclOutput
}

type PreSignedURLInput struct {
	HTTPMethod          enum.HttpMethodType
	Bucket              string
	Key                 string
	Expires             int64 // Expiration time in seconds, default 3600 seconds, max 7 days, range [1, 604800]
	Header              map[string]string
	Query               map[string]string
	AlternativeEndpoint string
}

type PreSignedURLOutput struct {
	SignedUrl    string            //  Pre-signed URL
	SignedHeader map[string]string // The actual header fields contained in the pre-signature
}

type CreateBucketInput struct {
	Bucket           string `json:"Bucket,omitempty"`           // required
	ACL              string `json:"ACL,omitempty"`              // optional
	GrantFullControl string `json:"GrantFullControl,omitempty"` // optional
	GrantRead        string `json:"GrantRead,omitempty"`        // optional
	GrantReadAcp     string `json:"GrantReadAcp,omitempty"`     // optional
	GrantWrite       string `json:"GrantWrite,omitempty"`       // optional
	GrantWriteAcp    string `json:"GrantWriteAcp,omitempty"`    // optional
}

type CreateBucketV2Input struct {
	Bucket           string                // required
	ACL              enum.ACLType          `location:"header" locationName:"X-Tos-Acl"`                // optional
	GrantFullControl string                `location:"header" locationName:"X-Tos-Grant-Full-Control"` // optional
	GrantRead        string                `location:"header" locationName:"X-Tos-Grant-Read"`         // optional
	GrantReadAcp     string                `location:"header" locationName:"X-Tos-Grant-Read-Acp"`     // optional
	GrantWrite       string                `location:"header" locationName:"X-Tos-Grant-Write"`        // optional
	GrantWriteAcp    string                `location:"header" locationName:"X-Tos-Grant-Write-Acp"`    // optional
	StorageClass     enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`      // setting the default storage type for buckets
	AzRedundancy     enum.AzRedundancyType `location:"header" locationName:"X-Tos-Az-Redundancy"`      // setting the AZ type for buckets
}

type CreateBucketOutput struct {
	RequestInfo `json:"-"`
	Location    string `json:"Location,omitempty"`
}

type CreateBucketV2Output struct {
	CreateBucketOutput
}

type HeadBucketOutput struct {
	RequestInfo  `json:"-"`
	Region       string                `json:"Region,omitempty"`
	StorageClass enum.StorageClassType `json:"StorageClass,omitempty"`
	AzRedundancy enum.AzRedundancyType `json:"AzRedundancy"`
}

type GetBucketCORSInput struct {
	Bucket string
}

type CorsRule struct {
	AllowedOrigin []string `json:"AllowedOrigins,omitempty"`
	AllowedMethod []string `json:"AllowedMethods,omitempty"`
	AllowedHeader []string `json:"AllowedHeaders,omitempty"`
	ExposeHeader  []string `json:"ExposeHeaders,omitempty"`
	MaxAgeSeconds int      `json:"MaxAgeSeconds,omitempty"`
}

type GetBucketCORSOutput struct {
	RequestInfo `json:"-"`
	CORSRules   []CorsRule `json:"CORSRules,omitempty"`
}

type PutBucketCORSInput struct {
	Bucket    string     `json:"-"`
	CORSRules []CorsRule `json:"CORSRules,omitempty"`
}

type PutBucketCORSOutput struct {
	RequestInfo `json:"-"`
}

type DeleteBucketCORSInput struct {
	Bucket string
}

type DeleteBucketCORSOutput struct {
	RequestInfo `json:"-"`
}

type HeadBucketInput struct {
	Bucket string
}

type DeleteBucketInput struct {
	Bucket string
}

type DeleteBucketOutput struct {
	RequestInfo `json:"-"`
}

type ListedOwner struct {
	ID string `json:"ID,omitempty"`
}

type ListBucketsOutput struct {
	RequestInfo `json:"-"`
	Buckets     []ListedBucket `json:"Buckets,omitempty"`
	Owner       ListedOwner    `json:"Owner,omitempty"`
}

type Owner struct {
	ID          string `json:"ID,omitempty"`
	DisplayName string `json:"DisplayName,omitempty"`
}

type ListedBucket struct {
	CreationDate     string `json:"CreationDate,omitempty"`
	Name             string `json:"Name,omitempty"`
	Location         string `json:"Location,omitempty"`
	ExtranetEndpoint string `json:"ExtranetEndpoint,omitempty"`
	IntranetEndpoint string `json:"IntranetEndpoint,omitempty"`
}

type ListBucketsInput struct{}

type PutObjectBasicInput struct {
	Bucket             string
	Key                string
	ContentLength      int64        `location:"header" locationName:"Content-Length"`
	ContentMD5         string       `location:"header" locationName:"Content-MD5"`
	ContentSHA256      string       `location:"header" locationName:"X-Tos-Content-Sha256"`
	CacheControl       string       `location:"header" locationName:"Cache-Control"`
	ContentDisposition string       `location:"header" locationName:"Content-Disposition" encodeChinese:"true"`
	ContentEncoding    string       `location:"header" locationName:"Content-Encoding"`
	ContentLanguage    string       `location:"header" locationName:"Content-Language"`
	ContentType        string       `location:"header" locationName:"Content-Type"`
	Expires            time.Time    `location:"header" locationName:"Expires"`
	ACL                enum.ACLType `location:"header" locationName:"X-Tos-Acl"`

	GrantFullControl string `location:"header" locationName:"X-Tos-Grant-Full-Control"` // optional
	GrantRead        string `location:"header" locationName:"X-Tos-Grant-Read"`         // optional
	GrantReadAcp     string `location:"header" locationName:"X-Tos-Grant-Read-Acp"`     // optional
	GrantWriteAcp    string `location:"header" locationName:"X-Tos-Grant-Write-Acp"`    // optional

	WebsiteRedirectLocation string                `location:"header" locationName:"X-Tos-Website-Redirect-Location"`
	StorageClass            enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
	SSECAlgorithm           string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey                 string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5              string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption    string                `location:"header" locationName:"X-Tos-Server-Side-Encryption"`
	Meta                    map[string]string     `location:"headers"`
	DataTransferListener    DataTransferListener
	RateLimiter             RateLimiter
}

type PutObjectV2Input struct {
	PutObjectBasicInput
	Content io.Reader
}

type PutObjectV2Output struct {
	RequestInfo
	ETag          string
	SSECAlgorithm string
	SSECKeyMD5    string
	VersionID     string
	HashCrc64ecma uint64
}

type PutObjectOutput struct {
	RequestInfo          `json:"-"`
	ETag                 string `json:"ETag,omitempty"`
	VersionID            string `json:"VersionId,omitempty"`
	SSECustomerAlgorithm string `json:"SSECustomerAlgorithm,omitempty"`
	SSECustomerKeyMD5    string `json:"SSECustomerKeyMD5,omitempty"`
}

type PutObjectFromFileInput struct {
	PutObjectBasicInput
	FilePath string
}

type PutObjectFromFileOutput struct {
	PutObjectV2Output
}

type CommonHeaders struct {
	ContentLength      int64        `location:"header" locationName:"Content-Length"`
	ContentMD5         string       `location:"header" locationName:"Content-MD5"`
	ContentSHA256      string       `location:"header" locationName:"X-Tos-Content-Sha256"`
	CacheControl       string       `location:"header" locationName:"Cache-Control"`
	ContentDisposition string       `location:"header" locationName:"Content-Disposition" encodeChinese:"true"`
	ContentEncoding    string       `location:"header" locationName:"Content-Encoding"`
	ContentLanguage    string       `location:"header" locationName:"Content-Language"`
	ContentType        string       `location:"header" locationName:"Content-Type"`
	Expires            time.Time    `location:"header" locationName:"Expires"`
	ACL                enum.ACLType `location:"header" locationName:"X-Tos-Acl"`

	GrantFullControl string `location:"header" locationName:"X-Tos-Grant-Full-Control"` // optional
	GrantRead        string `location:"header" locationName:"X-Tos-Grant-Read"`         // optional
	GrantReadAcp     string `location:"header" locationName:"X-Tos-Grant-Read-Acp"`     // optional
	GrantWriteAcp    string `location:"header" locationName:"X-Tos-Grant-Write-Acp"`    // optional

	WebsiteRedirectLocation string                `location:"header" locationName:"X-Tos-Website-Redirect-Location"`
	StorageClass            enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
}

type SSEHeaders struct {
	SSECAlgorithm        string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey              string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5           string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption string `location:"header" locationName:"X-Tos-Server-Side-Encryption"`
}

type AppendObjectV2Input struct {
	Bucket             string
	Key                string
	Offset             int64 `location:"query" locationName:"offset" default:"0"`
	Content            io.Reader
	ContentLength      int64        `location:"header" locationName:"Content-Length"`
	ContentMD5         string       `location:"header" locationName:"Content-MD5"`
	ContentSHA256      string       `location:"header" locationName:"X-Tos-Content-Sha256"`
	CacheControl       string       `location:"header" locationName:"Cache-Control"`
	ContentDisposition string       `location:"header" locationName:"Content-Disposition" encodeChinese:"true"`
	ContentEncoding    string       `location:"header" locationName:"Content-Encoding"`
	ContentLanguage    string       `location:"header" locationName:"Content-Language"`
	ContentType        string       `location:"header" locationName:"Content-Type"`
	Expires            time.Time    `location:"header" locationName:"Expires"`
	ACL                enum.ACLType `location:"header" locationName:"X-Tos-Acl"`

	GrantFullControl string `location:"header" locationName:"X-Tos-Grant-Full-Control"` // optional
	GrantRead        string `location:"header" locationName:"X-Tos-Grant-Read"`         // optional
	GrantReadAcp     string `location:"header" locationName:"X-Tos-Grant-Read-Acp"`     // optional
	GrantWriteAcp    string `location:"header" locationName:"X-Tos-Grant-Write-Acp"`    // optional

	WebsiteRedirectLocation string                `location:"header" locationName:"X-Tos-Website-Redirect-Location"`
	StorageClass            enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`

	Meta                 map[string]string `location:"headers"`
	DataTransferListener DataTransferListener
	RateLimiter          RateLimiter
	PreHashCrc64ecma     uint64
}

type AppendObjectOutput struct {
	RequestInfo      `json:"-"`
	ETag             string `json:"ETag,omitempty"`
	NextAppendOffset int64  `json:"NextAppendOffset,omitempty"`
}

type AppendObjectV2Output struct {
	RequestInfo      `json:"-"`
	VersionID        string `json:"VersionID,omitempty"`
	NextAppendOffset int64  `json:"NextAppendOffset,omitempty"`
	HashCrc64ecma    uint64 `json:"HashCrc64Ecma,omitempty"`
}

type SetObjectMetaInput struct {
	Bucket    string
	Key       string
	VersionID string

	CacheControl       string    `location:"header" locationName:"Cache-Control"`
	ContentDisposition string    `location:"header" locationName:"Content-Disposition"`
	ContentEncoding    string    `location:"header" locationName:"Content-Encoding"`
	ContentLanguage    string    `location:"header" locationName:"Content-Language"`
	ContentType        string    `location:"header" locationName:"Content-Type"`
	Expires            time.Time `location:"header" locationName:"Expires"`

	Meta map[string]string `location:"headers"`
}

type SetObjectMetaOutput struct {
	RequestInfo `json:"-"`
}

type ListObjectsV2Input struct {
	Bucket string
	ListObjectsInput
}

type ListObjectsInput struct {
	Prefix       string `location:"query" locationName:"prefix"`
	Delimiter    string `location:"query" locationName:"delimiter"`
	Marker       string `location:"query" locationName:"marker"`
	MaxKeys      int    `location:"query" locationName:"max-keys"`
	Reverse      bool   `location:"query" locationName:"reverse"`
	EncodingType string `location:"query" locationName:"encoding-type"` // "" or "url"
}

type ListedObject struct {
	Key          string `json:"Key,omitempty"`
	LastModified string `json:"LastModified,omitempty"`
	ETag         string `json:"ETag,omitempty"`
	Size         int64  `json:"Size,omitempty"`
	Owner        Owner  `json:"Owner,omitempty"`
	StorageClass string `json:"StorageClass,omitempty"`
	Type         string `json:"Type,omitempty"`
}

type ListedObjectV2 struct {
	Key           string
	LastModified  time.Time
	ETag          string
	Size          int64
	Owner         Owner
	StorageClass  enum.StorageClassType
	HashCrc64ecma uint64
}

type listedObjectV2 struct {
	Key           string
	LastModified  time.Time
	ETag          string
	Size          int64
	Owner         Owner
	StorageClass  enum.StorageClassType
	HashCrc64ecma string
}

type ListedCommonPrefix struct {
	Prefix string `json:"Prefix,omitempty"`
}

type ListObjectsOutput struct {
	RequestInfo    `json:"-"`
	Name           string               `json:"Name,omitempty"` // bucket name
	Prefix         string               `json:"Prefix,omitempty"`
	Marker         string               `json:"Marker,omitempty"`
	MaxKeys        int64                `json:"MaxKeys,omitempty"`
	NextMarker     string               `json:"NextMarker,omitempty"`
	Delimiter      string               `json:"Delimiter,omitempty"`
	IsTruncated    bool                 `json:"IsTruncated,omitempty"`
	EncodingType   string               `json:"EncodingType,omitempty"`
	CommonPrefixes []ListedCommonPrefix `json:"CommonPrefixes,omitempty"`
	Contents       []ListedObject       `json:"Contents,omitempty"`
}

type ListObjectsV2Output struct {
	RequestInfo    `json:"-"`
	Name           string               `json:"Name,omitempty"`
	Prefix         string               `json:"Prefix,omitempty"`
	Marker         string               `json:"Marker,omitempty"`
	MaxKeys        int64                `json:"MaxKeys,omitempty"`
	NextMarker     string               `json:"NextMarker,omitempty"`
	Delimiter      string               `json:"Delimiter,omitempty"`
	IsTruncated    bool                 `json:"IsTruncated,omitempty"`
	EncodingType   string               `json:"EncodingType,omitempty"`
	CommonPrefixes []ListedCommonPrefix `json:"CommonPrefixes,omitempty"`
	Contents       []ListedObjectV2     `json:"Contents,omitempty"`
}

type listObjectsV2Output struct {
	RequestInfo    `json:"-"`
	Name           string               `json:"Name,omitempty"`
	Prefix         string               `json:"Prefix,omitempty"`
	Marker         string               `json:"Marker,omitempty"`
	MaxKeys        int64                `json:"MaxKeys,omitempty"`
	NextMarker     string               `json:"NextMarker,omitempty"`
	Delimiter      string               `json:"Delimiter,omitempty"`
	IsTruncated    bool                 `json:"IsTruncated,omitempty"`
	EncodingType   string               `json:"EncodingType,omitempty"`
	CommonPrefixes []ListedCommonPrefix `json:"CommonPrefixes,omitempty"`
	Contents       []listedObjectV2     `json:"Contents,omitempty"`
}

type ListObjectVersionsInput struct {
	Prefix          string `location:"query" locationName:"prefix"`
	Delimiter       string `location:"query" locationName:"delimiter"`
	KeyMarker       string `location:"query" locationName:"key-marker"`
	VersionIDMarker string `location:"query" locationName:"version-id-marker"`
	MaxKeys         int    `location:"query" locationName:"max-keys"`
	EncodingType    string `location:"query" locationName:"encoding-type"` // "" or "url"
}

type ListObjectVersionsV2Input struct {
	Bucket string `json:"Prefix,omitempty"`
	ListObjectVersionsInput
}

type ListedObjectVersion struct {
	ETag         string `json:"ETag,omitempty"`
	IsLatest     bool   `json:"IsLatest,omitempty"`
	Key          string `json:"Key,omitempty"`
	LastModified string `json:"LastModified,omitempty"`
	Owner        Owner  `json:"Owner,omitempty"`
	Size         int64  `json:"Size,omitempty"`
	StorageClass string `json:"StorageClass,omitempty"`
	Type         string `json:"Type,omitempty"`
	VersionID    string `json:"VersionId,omitempty"`
}

type listedObjectVersionV2 struct {
	Key           string
	LastModified  time.Time
	ETag          string
	IsLatest      bool
	Size          int64
	Owner         Owner
	StorageClass  enum.StorageClassType
	VersionID     string
	HashCrc64ecma string
}

type ListedObjectVersionV2 struct {
	Key           string
	LastModified  time.Time
	ETag          string
	IsLatest      bool
	Size          int64
	Owner         Owner
	StorageClass  enum.StorageClassType
	VersionID     string
	HashCrc64ecma uint64
}

type ListedDeleteMarkerEntry struct {
	IsLatest     bool   `json:"IsLatest,omitempty"`
	Key          string `json:"Key,omitempty"`
	LastModified string `json:"LastModified,omitempty"`
	Owner        Owner  `json:"Owner,omitempty"`
	VersionID    string `json:"VersionId,omitempty"`
}

type ListedDeleteMarker struct {
	Key          string
	LastModified time.Time
	IsLatest     bool
	Owner        Owner
	VersionID    string
}

type listObjectVersionsV2Output struct {
	RequestInfo         `json:"-"`
	Name                string                  `json:"Name,omitempty"` // bucket name
	Prefix              string                  `json:"Prefix,omitempty"`
	KeyMarker           string                  `json:"KeyMarker,omitempty"`
	VersionIDMarker     string                  `json:"VersionIdMarker,omitempty"`
	Delimiter           string                  `json:"Delimiter,omitempty"`
	EncodingType        string                  `json:"EncodingType,omitempty"`
	MaxKeys             int                     `json:"MaxKeys,omitempty"`
	NextKeyMarker       string                  `json:"NextKeyMarker,omitempty"`
	NextVersionIDMarker string                  `json:"NextVersionIdMarker,omitempty"`
	IsTruncated         bool                    `json:"IsTruncated,omitempty"`
	CommonPrefixes      []ListedCommonPrefix    `json:"CommonPrefixes,omitempty"`
	Versions            []listedObjectVersionV2 `json:"Versions,omitempty"`
	DeleteMarkers       []ListedDeleteMarker    `json:"DeleteMarkers,omitempty"`
}

type ListObjectVersionsV2Output struct {
	RequestInfo         `json:"-"`
	Name                string                  `json:"Name,omitempty"` // bucket name
	Prefix              string                  `json:"Prefix,omitempty"`
	KeyMarker           string                  `json:"KeyMarker,omitempty"`
	VersionIDMarker     string                  `json:"VersionIdMarker,omitempty"`
	Delimiter           string                  `json:"Delimiter,omitempty"`
	EncodingType        string                  `json:"EncodingType,omitempty"`
	MaxKeys             int                     `json:"MaxKeys,omitempty"`
	NextKeyMarker       string                  `json:"NextKeyMarker,omitempty"`
	NextVersionIDMarker string                  `json:"NextVersionIdMarker,omitempty"`
	IsTruncated         bool                    `json:"IsTruncated,omitempty"`
	CommonPrefixes      []ListedCommonPrefix    `json:"CommonPrefixes,omitempty"`
	Versions            []ListedObjectVersionV2 `json:"Versions,omitempty"`
	DeleteMarkers       []ListedDeleteMarker    `json:"DeleteMarkers,omitempty"`
}

type ListObjectVersionsOutput struct {
	RequestInfo         `json:"-"`
	Name                string                    `json:"Name,omitempty"` // bucket name
	Prefix              string                    `json:"Prefix,omitempty"`
	KeyMarker           string                    `json:"KeyMarker,omitempty"`
	VersionIDMarker     string                    `json:"VersionIdMarker,omitempty"`
	Delimiter           string                    `json:"Delimiter,omitempty"`
	EncodingType        string                    `json:"EncodingType,omitempty"`
	MaxKeys             int64                     `json:"MaxKeys,omitempty"`
	NextKeyMarker       string                    `json:"NextKeyMarker,omitempty"`
	NextVersionIDMarker string                    `json:"NextVersionIdMarker,omitempty"`
	IsTruncated         bool                      `json:"IsTruncated,omitempty"`
	CommonPrefixes      []ListedCommonPrefix      `json:"CommonPrefixes,omitempty"`
	Versions            []ListedObjectVersion     `json:"Versions,omitempty"`
	DeleteMarkers       []ListedDeleteMarkerEntry `json:"DeleteMarkers,omitempty"`
}

type GetObjectOutput struct {
	RequestInfo  `json:"-"`
	ContentRange string        `json:"ContentRange,omitempty"`
	Content      io.ReadCloser `json:"-"`
	ObjectMeta
}

type GetObjectV2Input struct {
	Bucket     string
	Key        string
	VersionID  string `location:"query" locationName:"versionId"`
	PartNumber int    `location:"query" locationName:"partNumber"`

	IfMatch           string    `location:"header" locationName:"If-Match"`
	IfModifiedSince   time.Time `location:"header" locationName:"If-Modified-Since"`
	IfNoneMatch       string    `location:"header" locationName:"If-None-Match"`
	IfUnmodifiedSince time.Time `location:"header" locationName:"If-Unmodified-Since"`

	SSECAlgorithm string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey       string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5    string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`

	ResponseCacheControl       string    `location:"query" locationName:"Cache-Control"`
	ResponseContentDisposition string    `location:"query" locationName:"Content-Disposition"`
	ResponseContentEncoding    string    `location:"query" locationName:"Content-Encoding"`
	ResponseContentLanguage    string    `location:"query" locationName:"Content-Language"`
	ResponseContentType        string    `location:"query" locationName:"Content-Type"`
	ResponseExpires            time.Time `location:"query" locationName:"Expires"`

	RangeStart int64
	RangeEnd   int64

	DataTransferListener DataTransferListener
	RateLimiter          RateLimiter
}

type GetObjectBasicOutput struct {
	RequestInfo
	ContentRange string // don't move into ObjectMetaV2
	ObjectMetaV2
}

type GetObjectV2Output struct {
	GetObjectBasicOutput
	Content io.ReadCloser
}

type GetObjectToFileInput struct {
	GetObjectV2Input
	FilePath string
}

type GetObjectToFileOutput struct {
	GetObjectBasicOutput
}

type HeadObjectV2Input struct {
	Bucket    string
	Key       string
	VersionID string `location:"query" locationName:"versionId"`

	IfMatch           string    `location:"header" locationName:"If-Match"`
	IfModifiedSince   time.Time `location:"header" locationName:"If-Modified-Since"`
	IfNoneMatch       string    `location:"header" locationName:"If-None-Match"`
	IfUnmodifiedSince time.Time `location:"header" locationName:"If-Unmodified-Since"`

	SSECAlgorithm string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey       string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5    string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
}

type HeadObjectOutput struct {
	RequestInfo  `json:"-"`
	ContentRange string `json:"ContentRange,omitempty"`
	ObjectMeta
}

type HeadObjectV2Output struct {
	RequestInfo `json:"-"`
	ObjectMetaV2
}

type DeleteObjectV2Input struct {
	Bucket    string
	Key       string
	VersionID string `location:"query" locationName:"versionId"`
}

type DeleteObjectOutput struct {
	RequestInfo  `json:"-"`
	DeleteMarker bool   `json:"DeleteMarker,omitempty"`
	VersionID    string `json:"VersionId,omitempty"`
}

type DeleteObjectV2Output struct {
	DeleteObjectOutput
}

type ObjectTobeDeleted struct {
	Key       string `json:"Key,omitempty"`
	VersionID string `json:"VersionId,omitempty"`
}

type DeleteMultiObjectsInput struct {
	Bucket  string
	Objects []ObjectTobeDeleted `json:"Objects,omitempty"`
	Quiet   bool                `json:"Quiet,omitempty"`
}

type Deleted struct {
	Key                   string `json:"Key,omitempty"`
	VersionID             string `json:"VersionId,omitempty"`
	DeleteMarker          *bool  `json:"DeleteMarker,omitempty"`
	DeleteMarkerVersionID string `json:"DeleteMarkerVersionId,omitempty"`
}

type DeletedV2 struct {
	Key                   string `json:"Key,omitempty"`
	VersionID             string `json:"VersionId,omitempty"`
	DeleteMarker          bool   `json:"DeleteMarker,omitempty"`
	DeleteMarkerVersionID string `json:"DeleteMarkerVersionId,omitempty"`
}

type DeleteError struct {
	Code      string `json:"Code,omitempty"`
	Message   string `json:"Message,omitempty"`
	Key       string `json:"Key,omitempty"`
	VersionID string `json:"VersionId,omitempty"`
}

type DeleteMultiObjectsOutput struct {
	RequestInfo `json:"-"`
	Deleted     []DeletedV2   `json:"Deleted,omitempty"` // 删除成功的Object列表
	Error       []DeleteError `json:"Error,omitempty"`   // 删除失败的Object列表
}

type CopyObjectInput struct {
	Bucket             string
	Key                string
	SrcBucket          string
	SrcKey             string
	SrcVersionID       string
	CacheControl       string       `location:"header" locationName:"Cache-Control"`
	ContentDisposition string       `location:"header" locationName:"Content-Disposition" encodeChinese:"true"`
	ContentEncoding    string       `location:"header" locationName:"Content-Encoding"`
	ContentLanguage    string       `location:"header" locationName:"Content-Language"`
	ContentType        string       `location:"header" locationName:"Content-Type"`
	Expires            time.Time    `location:"header" locationName:"Expires"`
	ACL                enum.ACLType `location:"header" locationName:"X-Tos-Acl"`

	GrantFullControl string `location:"header" locationName:"X-Tos-Grant-Full-Control"` // optional
	GrantRead        string `location:"header" locationName:"X-Tos-Grant-Read"`         // optional
	GrantReadAcp     string `location:"header" locationName:"X-Tos-Grant-Read-Acp"`     // optional
	GrantWriteAcp    string `location:"header" locationName:"X-Tos-Grant-Write-Acp"`    // optional

	WebsiteRedirectLocation string                `location:"header" locationName:"X-Tos-Website-Redirect-Location"`
	StorageClass            enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`

	CopySourceIfMatch           string    `location:"header" locationName:"X-Tos-Copy-Source-If-Match"`
	CopySourceIfModifiedSince   time.Time `location:"header" locationName:"X-Tos-Copy-Source-If-Modified-Since"`
	CopySourceIfNoneMatch       string    `location:"header" locationName:"X-Tos-Copy-Source-If-None-Match"`
	CopySourceIfUnmodifiedSince time.Time `location:"header" locationName:"X-Tos-Copy-Source-If-Unmodified-Since"`

	CopySourceSSECAlgorithm string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	CopySourceSSECKey       string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	CopySourceSSECKeyMD5    string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption    string `location:"header" locationName:"X-Tos-Server-Side-Encryption"`

	MetadataDirective enum.MetadataDirectiveType `location:"header" locationName:"X-Tos-Metadata-Directive"`
	Meta              map[string]string          `location:"headers"`
}

type CopyObjectOutput struct {
	RequestInfo     `json:"-"`
	VersionID       string `json:"VersionId,omitempty"`
	SourceVersionID string `json:"SourceVersionId,omitempty"`
	ETag            string `json:"ETag,omitempty"`         // at body
	LastModified    string `json:"LastModified,omitempty"` // at body
}

type UploadPartCopyInput struct {
	UploadID        string `json:"UploadId,omitempty"`
	DestinationKey  string `json:"DestinationKey,omitempty"`
	SourceBucket    string `json:"SourceBucket,omitempty"`
	SourceKey       string `json:"SourceKey,omitempty"`
	SourceVersionID string `json:"SourceVersionId,omitempty"` // optional
	StartOffset     *int64 `json:"StartOffset,omitempty"`     // optional
	PartSize        *int64 `json:"PartSize,omitempty"`        // optional
	PartNumber      int    `json:"PartNumber,omitempty"`
}

type UploadPartCopyOutput struct {
	RequestInfo     `json:"-"`
	VersionID       string `json:"VersionId,omitempty"`
	SourceVersionID string `json:"SourceVersionId,omitempty"`
	PartNumber      int    `json:"PartNumber,omitempty"`
	ETag            string `json:"ETag,omitempty"`
	LastModified    string `json:"LastModified,omitempty"`
}

type UploadPartCopyV2Input struct {
	Bucket     string
	Key        string
	UploadID   string `location:"query" locationName:"uploadId"`
	PartNumber int    `location:"query" locationName:"partNumber"`

	SrcBucket            string
	SrcKey               string
	SrcVersionID         string `location:"query" locationName:"versionId"`
	CopySourceRangeStart int64
	CopySourceRangeEnd   int64

	CopySourceIfMatch           string    `location:"header" locationName:"X-Tos-Copy-Source-If-Match"`
	CopySourceIfModifiedSince   time.Time `location:"header" locationName:"X-Tos-Copy-Source-If-Modified-Since"`
	CopySourceIfNoneMatch       string    `location:"header" locationName:"X-Tos-Copy-Source-If-None-Match"`
	CopySourceIfUnmodifiedSince time.Time `location:"header" locationName:"X-Tos-Copy-Source-If-Unmodified-Since"`

	CopySourceSSECAlgorithm string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	CopySourceSSECKey       string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	CopySourceSSECKeyMD5    string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
}

type UploadPartCopyV2Output struct {
	RequestInfo
	PartNumber          int
	ETag                string
	LastModified        time.Time
	CopySourceVersionID string
}

type CreateMultipartUploadV2Input struct {
	Bucket             string
	Key                string
	EncodingType       string       `location:"query" locationName:"encoding-type"` // "" or "url"
	CacheControl       string       `location:"header" locationName:"Cache-Control"`
	ContentDisposition string       `location:"header" locationName:"Content-Disposition" encodeChinese:"true"`
	ContentEncoding    string       `location:"header" locationName:"Content-Encoding"`
	ContentLanguage    string       `location:"header" locationName:"Content-Language"`
	ContentType        string       `location:"header" locationName:"Content-Type"`
	Expires            time.Time    `location:"header" locationName:"Expires"`
	ACL                enum.ACLType `location:"header" locationName:"X-Tos-Acl"`

	GrantFullControl string `location:"header" locationName:"X-Tos-Grant-Full-Control"` // optional
	GrantRead        string `location:"header" locationName:"X-Tos-Grant-Read"`         // optional
	GrantReadAcp     string `location:"header" locationName:"X-Tos-Grant-Read-Acp"`     // optional
	GrantWriteAcp    string `location:"header" locationName:"X-Tos-Grant-Write-Acp"`    // optional

	WebsiteRedirectLocation string                `location:"header" locationName:"X-Tos-Website-Redirect-Location"`
	StorageClass            enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
	SSECAlgorithm           string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey                 string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5              string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption    string                `location:"header" locationName:"X-Tos-Server-Side-Encryption"`
	Meta                    map[string]string     `location:"headers"`
}

type CreateMultipartUploadOutput struct {
	RequestInfo          `json:"-"`
	Bucket               string `json:"Bucket,omitempty"`
	Key                  string `json:"Key,omitempty"`
	UploadID             string `json:"UploadId,omitempty"`
	SSECustomerAlgorithm string `json:"SSECustomerAlgorithm,omitempty"`
	SSECustomerKeyMD5    string `json:"SSECustomerKeyMD5,omitempty"`
}

type CreateMultipartUploadV2Output struct {
	RequestInfo   `json:"-"`
	Bucket        string `json:"Bucket,omitempty"`
	Key           string `json:"Key,omitempty"`
	UploadID      string `json:"UploadID,omitempty"`
	SSECAlgorithm string `json:"SSECAlgorithm,omitempty"`
	SSECKeyMD5    string `json:"SSECKeyMD5,omitempty"`
	EncodingType  string `json:"EncodingType,omitempty"`
}

type UploadPartInput struct {
	Key        string    `json:"Key,omitempty"`
	UploadID   string    `json:"UploadId,omitempty"`
	PartNumber int       `json:"PartNumber,omitempty"`
	Content    io.Reader `json:"-"`
}

type UploadPartOutput struct {
	RequestInfo          `json:"-"`
	PartNumber           int    `json:"PartNumber,omitempty"`
	ETag                 string `json:"ETag,omitempty"`
	SSECustomerAlgorithm string `json:"SSECustomerAlgorithm,omitempty"`
	SSECustomerKeyMD5    string `json:"SSECustomerKeyMD5,omitempty"`
}

func (up *UploadPartOutput) uploadedPart() uploadedPart {
	return uploadedPart{PartNumber: up.PartNumber, ETag: up.ETag}
}

type UploadPartBasicInput struct {
	Bucket     string
	Key        string
	UploadID   string `location:"query" locationName:"uploadId"`
	PartNumber int    `location:"query" locationName:"partNumber"`

	ContentMD5 string `location:"header" locationName:"Content-MD5"`

	SSECAlgorithm        string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey              string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5           string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption string `location:"header" locationName:"X-Tos-Server-Side-Encryption"`

	DataTransferListener DataTransferListener
	RateLimiter          RateLimiter
}

type UploadPartV2Input struct {
	UploadPartBasicInput
	Content       io.Reader
	ContentLength int64 `location:"header" locationName:"Content-Length"`
}

type UploadPartV2Output struct {
	RequestInfo
	PartNumber    int
	ETag          string
	SSECAlgorithm string
	SSECKeyMD5    string
	HashCrc64ecma uint64
}

func (up *UploadPartV2Output) uploadedPart() uploadedPart {
	return uploadedPart{PartNumber: up.PartNumber, ETag: up.ETag}
}

type UploadPartFromFileInput struct {
	UploadPartBasicInput
	FilePath string
	Offset   uint64 // 当前分段在文件中的起始位置
	PartSize int64  // 当前分段长度，该字段等同于 Content-Length 头域
}

type UploadPartFromFileOutput struct {
	UploadPartV2Output
}

type UploadedPart struct {
	PartNumber   int32  `json:"PartNumber,omitempty"`   // Part编号
	ETag         string `json:"ETag,omitempty"`         // ETag
	LastModified string `json:"LastModified,omitempty"` // 最后一次修改时间
	Size         int64  `json:"Size,omitempty"`         // Part大小
}

func (part *UploadedPart) uploadedPart() uploadedPart {
	return uploadedPart{
		PartNumber: int(part.PartNumber),
		ETag:       part.ETag,
	}
}

type UploadedPartV2 struct {
	PartNumber   int       `json:"PartNumber,omitempty"`   // Part编号
	ETag         string    `json:"ETag,omitempty"`         // ETag
	LastModified time.Time `json:"LastModified,omitempty"` // 最后一次修改时间
	Size         int64     `json:"Size,omitempty"`         // Part大小
}

func (part UploadedPartV2) uploadedPart() uploadedPart {
	return uploadedPart{PartNumber: part.PartNumber, ETag: part.ETag}
}

type MultipartUploadedPart interface {
	uploadedPart() uploadedPart
}

type CompleteMultipartUploadInput struct {
	Key           string                  `json:"Key,omitempty"`
	UploadID      string                  `json:"UploadId,omitempty"`
	UploadedParts []MultipartUploadedPart `json:"UploadedParts,omitempty"`
}

type CompleteMultipartUploadOutput struct {
	RequestInfo `json:"-"`
	VersionID   string `json:"VersionId,omitempty"`
}

type CompleteMultipartUploadV2Input struct {
	Bucket   string
	Key      string
	UploadID string `location:"query" locationName:"uploadId"`
	Parts    []UploadedPartV2
}

type CompleteMultipartUploadV2Output struct {
	RequestInfo
	Bucket        string
	Key           string
	ETag          string
	Location      string
	VersionID     string
	HashCrc64ecma uint64
}

type AbortMultipartUploadInput struct {
	// Bucket is needed in V2 api
	Bucket   string
	Key      string
	UploadID string `location:"query" locationName:"uploadId"`
}

type AbortMultipartUploadOutput struct {
	RequestInfo `json:"-"`
}

type UploadInfo struct {
	Key          string `json:"Key,omitempty"`
	UploadId     string `json:"UploadId,omitempty"`
	Owner        Owner  `json:"Owner,omitempty"`
	StorageClass string `json:"StorageClass,omitempty"`
	Initiated    string `json:"Initiated,omitempty"`
}

type UploadCommonPrefix struct {
	Prefix string `json:"Prefix"`
}

type ListMultipartUploadsInput struct {
	Prefix         string `json:"Prefix,omitempty"`
	Delimiter      string `json:"Delimiter,omitempty"`
	KeyMarker      string `json:"KeyMarker,omitempty"`
	UploadIDMarker string `json:"UploadIdMarker,omitempty"`
	MaxUploads     int    `json:"MaxUploads,omitempty"`
}

type ListMultipartUploadsOutput struct {
	RequestInfo        `json:"-"`
	Bucket             string               `json:"Bucket,omitempty"`
	KeyMarker          string               `json:"KeyMarker,omitempty"`
	UploadIdMarker     string               `json:"UploadIdMarker,omitempty"`
	NextKeyMarker      string               `json:"NextKeyMarker,omitempty"`
	NextUploadIdMarker string               `json:"NextUploadIdMarker,omitempty"`
	Delimiter          string               `json:"Delimiter,omitempty"`
	Prefix             string               `json:"Prefix,omitempty"`
	MaxUploads         int32                `json:"MaxUploads,omitempty"`
	IsTruncated        bool                 `json:"IsTruncated,omitempty"`
	Upload             []UploadInfo         `json:"Uploads,omitempty"`
	CommonPrefixes     []UploadCommonPrefix `json:"CommonPrefixes,omitempty"`
}

type ListMultipartUploadsV2Input struct {
	Bucket         string
	Prefix         string `location:"query" locationName:"uploads"`
	Delimiter      string `location:"query" locationName:"delimiter"`
	KeyMarker      string `location:"query" locationName:"key-marker"`
	UploadIDMarker string `location:"query" locationName:"upload-id-marker"`
	MaxUploads     int    `location:"query" locationName:"max-uploads"`
	EncodingType   string `location:"query" locationName:"encoding-type"` // "" or "url"
}

type ListedUpload struct {
	Key          string
	UploadID     string
	Owner        Owner
	StorageClass enum.StorageClassType
	Initiated    time.Time
}

type ListMultipartUploadsV2Output struct {
	RequestInfo
	Bucket             string
	Prefix             string
	KeyMarker          string
	UploadIDMarker     string
	MaxUploads         int
	Delimiter          string
	IsTruncated        bool
	EncodingType       string
	NextKeyMarker      string
	NextUploadIDMarker string
	CommonPrefixes     []ListedCommonPrefix
	Uploads            []ListedUpload
}

type ListUploadedPartsInput struct {
	Key              string `json:"Key,omitempty"`
	UploadID         string `json:"UploadId,omitempty"`
	MaxParts         int    `json:"MaxParts,omitempty"`             // 最大Part个数
	PartNumberMarker int    `json:"NextPartNumberMarker,omitempty"` // 起始Part的位置
}

type ListUploadedPartsOutput struct {
	RequestInfo          `json:"-"`
	Bucket               string         `json:"Bucket,omitempty"`               // Bucket名称
	Key                  string         `json:"Key,omitempty"`                  // Object名称
	UploadID             string         `json:"UploadId,omitempty"`             // 上传ID
	PartNumberMarker     int            `json:"PartNumberMarker,omitempty"`     // 当前页起始位置
	NextPartNumberMarker int            `json:"NextPartNumberMarker,omitempty"` // 下一个Part的位置
	MaxParts             int            `json:"MaxParts,omitempty"`             // 最大Part个数
	IsTruncated          bool           `json:"IsTruncated,omitempty"`          // 是否完全上传完成
	StorageClass         string         `json:"StorageClass,omitempty"`         // 存储类型
	Owner                Owner          `json:"Owner,omitempty"`                // 属主
	UploadedParts        []UploadedPart `json:"Parts,omitempty"`                // 已完成的Part
}

type ListPartsInput struct {
	Bucket           string
	Key              string
	UploadID         string `location:"query" locationName:"uploadId"`
	PartNumberMarker int    `location:"query" locationName:"part-number-marker"`
	MaxParts         int    `location:"query" locationName:"max-parts"`
	EncodingType     string `location:"query" locationName:"encoding-type"` // "" or "url"
}

type ListPartsOutput struct {
	RequestInfo
	Bucket           string
	Key              string
	UploadID         string
	PartNumberMarker int
	MaxParts         int
	IsTruncated      bool
	EncodingType     string

	NextPartNumberMarker int
	StorageClass         enum.StorageClassType
	Owner                Owner
	Parts                []UploadedPartV2
}

type CancelHook interface {
	// Cancel 取消断点上传\断点下载事, isAbort 为 true 时删除上下文信息和临时文件，为 false 时只是中断当前执行，该接口只能调用一次
	Cancel(isAbort bool)
	// to make user unable to implement this interface
	internal()
}

type DownloadFileInput struct {
	HeadObjectV2Input
	FilePath              string
	PartSize              int64
	TaskNum               int
	EnableCheckpoint      bool
	CheckpointFile        string
	tempFile              string
	DownloadEventListener DownloadEventListener
	DataTransferListener  DataTransferListener
	RateLimiter           RateLimiter
	CancelHook            CancelHook // user can not set this filed
}

func (d *DownloadFileInput) withCancelHook(hook CancelHook) {
	d.CancelHook = hook
}

type DownloadFileOutput struct {
	HeadObjectV2Output
}

type DownloadEvent struct {
	Type           enum.DownloadEventType
	Err            error // not empty when it occurs when failed, aborted event occurs
	Bucket         string
	Key            string
	VersionID      string
	FilePath       string  // path of the file to download to
	CheckpointFile *string // path to checkpoint file
	TempFilePath   *string // path fo the temp file
	// not empty when download part event occurs
	DowloadPartInfo *DownloadPartInfo
}

// DownloadPartInfo is returned when DownloadEvent occur
type DownloadPartInfo struct {
	PartNumber int
	RangeStart int64
	RangeEnd   int64
}

type DownloadEventListener interface {
	EventChange(event *DownloadEvent)
}

type UploadFileInput struct {
	CreateMultipartUploadV2Input

	FilePath             string
	PartSize             int64
	TaskNum              int
	EnableCheckpoint     bool
	CheckpointFile       string
	DataTransferListener DataTransferListener
	UploadEventListener  UploadEventListener
	RateLimiter          RateLimiter
	// cancelHook 支持取消断点续传任务
	CancelHook CancelHook
}

func NewCancelHook() CancelHook {
	return &canceler{
		cancelHandle: make(chan struct{}),
	}
}

// UploadPartInfo is returned when UploadEvent occur
type UploadPartInfo struct {
	PartNumber int
	PartSize   int64
	Offset     int64
	// upload part succeed 事件发生时有值
	ETag          *string
	HashCrc64ecma *uint64
}

type UploadEvent struct {
	Type           enum.UploadEventType
	Err            error // failed, aborted 事件发生时不为空
	Bucket         string
	Key            string
	UploadID       *string
	CheckpointFile *string // 断点续传文件全路径
	// upload part 相关事件发生时有值
	UploadPartInfo *UploadPartInfo
}

type UploadEventListener interface {
	EventChange(event *UploadEvent)
}

type UploadFileOutput struct {
	RequestInfo
	Bucket        string
	Key           string
	UploadID      string
	ETag          string
	Location      string
	VersionID     string
	HashCrc64ecma uint64
	SSECAlgorithm string
	SSECKeyMD5    string
	EncodingType  string
}

type DataTransferStatus struct {
	TotalBytes    int64
	ConsumedBytes int64 // bytes read/written
	RWOnceBytes   int64 // bytes read/written this time
	Type          enum.DataTransferType
}

type DataTransferListener interface {
	DataTransferStatusChange(status *DataTransferStatus)
}

type RateLimiter interface {
	// Acquire try to get a token.
	// If ok, caller can read want bytes, else wait timeToWait and try again.
	Acquire(want int64) (ok bool, timeToWait time.Duration)
}
