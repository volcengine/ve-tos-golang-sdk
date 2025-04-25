package tos

import (
	"fmt"
	"io"
	"net/url"
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
	ID          string           `json:"ID,omitempty"`
	DisplayName string           `json:"DisplayName,omitempty"`
	Type        enum.GranteeType `json:"Type,omitempty"`
	Canned      enum.CannedType  `json:"Canned,omitempty"`
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
	Owner                Owner   `json:"Owner,omitempty"`
	Grants               []Grant `json:"Grants,omitempty"`
	IsDefault            bool    `json:"IsDefault,omitempty"`
	BucketOwnerEntrusted bool    `json:"BucketOwnerEntrusted,omitempty"`
}

type GetObjectAclOutput struct {
	RequestInfo          `json:"-"`
	VersionID            string  `json:"VersionId,omitempty"`
	Owner                Owner   `json:"Owner,omitempty"`
	Grants               []Grant `json:"Grants,omitempty"`
	BucketOwnerEntrusted bool    `json:"BucketOwnerEntrusted,omitempty"`
	IsDefault            bool    `json:"IsDefault,omitempty"`
}

type bucketACL struct {
	Owner              Owner     `json:"Owner,omitempty"`
	GrantList          []GrantV2 `json:"Grants,omitempty"`
	BucketAclDelivered bool      `json:"BucketAclDelivered"`
}

type PutBucketACLInput struct {
	Bucket           string
	ACLType          enum.ACLType `location:"header" locationName:"X-Tos-Acl"`                // optional
	GrantFullControl string       `location:"header" locationName:"X-Tos-Grant-Full-Control"` // optional
	GrantRead        string       `location:"header" locationName:"X-Tos-Grant-Read"`         // optional
	GrantReadAcp     string       `location:"header" locationName:"X-Tos-Grant-Read-Acp"`     // optional
	GrantWrite       string       `location:"header" locationName:"X-Tos-Grant-Write"`        // optional
	GrantWriteAcp    string       `location:"header" locationName:"X-Tos-Grant-Write-Acp"`    // optional

	Owner              Owner     `json:"Owner,omitempty"`
	Grants             []GrantV2 `json:"Grants,omitempty"`
	BucketAclDelivered bool      `json:"BucketAclDelivered,omitempty"`
}

type PutBucketACLOutput struct {
	RequestInfo
}

type GetBucketACLInput struct {
	Bucket string
}

type GetBucketACLOutput struct {
	RequestInfo
	Owner              Owner     `json:"Owner,omitempty"`
	Grants             []GrantV2 `json:"Grants,omitempty"`
	BucketAclDelivered bool      `json:"BucketAclDelivered"`
}

type GetObjectACLInput struct {
	Bucket    string
	Key       string
	VersionID string `location:"query" locationName:"versionId"`
}

type GetObjectACLOutput struct {
	RequestInfo          `json:"-"`
	VersionID            string    `json:"VersionID,omitempty"`
	Owner                Owner     `json:"Owner,omitempty"`
	Grants               []GrantV2 `json:"Grants,omitempty"`
	BucketOwnerEntrusted bool      `json:"BucketOwnerEntrusted"`
	IsDefault            bool      `json:"IsDefault"`
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

	// Deprecated
	GrantWrite           string `location:"header" locationName:"X-Tos-Grant-Write"`     // optional
	GrantWriteAcp        string `location:"header" locationName:"X-Tos-Grant-Write-Acp"` // optional
	Owner                Owner
	Grants               []GrantV2
	BucketOwnerEntrusted bool
	IsDefault            bool
}

type PutObjectAclOutput struct {
	RequestInfo `json:"-"`
}

type PutObjectACLOutput struct {
	PutObjectAclOutput
}

type putFetchTaskV2Input struct {
	URL              string `json:"URL,omitempty"`
	IgnoreSameKey    bool   `json:"IgnoreSameKey,omitempty"`
	ContentMD5       string `json:"ContentMD5,omitempty"`
	Object           string `json:"Object,omitempty"`
	CallbackUrl      string `json:"CallbackUrl,omitempty"`
	CallbackHost     string `json:"CallbackHost,omitempty"`
	CallbackBodyType string `json:"CallbackBodyType,omitempty"`
	CallbackBody     string `json:"CallbackBody,omitempty"`
}

type PutFetchTaskInputV2 struct {
	Bucket string
	Key    string

	ACL              enum.ACLType          `location:"header" locationName:"X-Tos-Acl"`
	GrantFullControl string                `location:"header" locationName:"X-Tos-Grant-Full-Control"`
	GrantRead        string                `location:"header" locationName:"X-Tos-Grant-Read"`
	GrantReadACP     string                `location:"header" locationName:"X-Tos-Grant-Read-Acp"`
	GrantWriteACP    string                `location:"header" locationName:"X-Tos-Grant-Write-Acp"`
	StorageClass     enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
	SSECAlgorithm    string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey          string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5       string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	Meta             map[string]string     `location:"headers"`

	URL           string `json:"URL,omitempty"`
	IgnoreSameKey bool   `json:"IgnoreSameKey,omitempty"`
	// Deprecated: use content md5 instead
	HexMD5 string

	ContentMD5       string `json:"ContentMD5,omitempty"`
	CallbackUrl      string `json:"CallbackUrl,omitempty"`
	CallbackHost     string `json:"CallbackHost,omitempty"`
	CallbackBodyType string `json:"CallbackBodyType,omitempty"`
	CallbackBody     string `json:"CallbackBody,omitempty"`
}

type PutFetchTaskOutputV2 struct {
	RequestInfo
	TaskID string
}

type FetchObjectInputV2 struct {
	Bucket           string
	Key              string
	ACL              enum.ACLType          `location:"header" locationName:"X-Tos-Acl"`
	GrantFullControl string                `location:"header" locationName:"X-Tos-Grant-Full-Control"`
	GrantRead        string                `location:"header" locationName:"X-Tos-Grant-Read"`
	GrantReadACP     string                `location:"header" locationName:"X-Tos-Grant-Read-Acp"`
	GrantWriteACP    string                `location:"header" locationName:"X-Tos-Grant-Write-Acp"`
	StorageClass     enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
	SSECAlgorithm    string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey          string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5       string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	Meta             map[string]string     `location:"headers"`

	URL           string `json:"URL,omitempty"`
	IgnoreSameKey bool   `json:"IgnoreSameKey,omitempty"`

	ContentMD5 string `json:"ContentMD5,omitempty"`

	// Deprecated: use content md5 instead
	HexMD5 string
}

type FetchObjectOutputV2 struct {
	RequestInfo
	VersionID     string `json:"VersionId,omitempty"`
	Etag          string `json:"Etag,omitempty"`
	SSECAlgorithm string `json:"SSECAlgorithm,omitempty"`
	SSECKeyMD5    string `json:"SSECKeyMD5,omitempty"`
}

type PreSingedPostSignatureInput struct {
	Bucket             string
	Key                string
	Expires            int64
	Conditions         []PostSignatureCondition
	ContentLengthRange *ContentLengthRange
}

type PreSingedPostSignatureOutput struct {
	OriginPolicy string
	Policy       string
	Algorithm    string
	Credential   string
	Date         string
	Signature    string
}

type ContentLengthRange struct {
	RangeStart int64
	RangeEnd   int64
}

type PostSignatureCondition struct {
	Key      string
	Value    string
	Operator *string
}

type PreSingedPolicyURLInput struct {
	Bucket              string
	Expires             int64
	Conditions          []PolicySignatureCondition
	AlternativeEndpoint string
	IsCustomDomain      bool
}

type PreSingedPolicyURLOutput struct {
	PreSignedPolicyURLGenerator
	SignatureQuery string
	bucket         string
	host           string
	scheme         string
	isCustomDomain bool
}

type PolicySignatureCondition struct {
	Key      string
	Value    string
	Operator *string
}

type PreSignedPolicyURLGenerator interface {
	GetSignedURLForList(bucket string, additionalQuery map[string]string) string
	GetSignedURLForGetOrHead(bucket, key string, additionalQuery map[string]string) string
}

func (output *PreSingedPolicyURLOutput) GetSignedURLForList(additionalQuery map[string]string) string {
	query := make(url.Values)
	for k, v := range additionalQuery {
		query.Add(k, v)
	}
	queryStr := query.Encode()
	if queryStr != "" {
		queryStr = "&" + queryStr
	}
	var domain string
	if output.isCustomDomain {
		domain = output.host
	} else {
		domain = fmt.Sprintf("%s.%s", output.bucket, output.host)
	}
	str := fmt.Sprintf("%s://%s/?%s%s", output.scheme, domain, output.SignatureQuery, queryStr)
	return str
}
func (output *PreSingedPolicyURLOutput) GetSignedURLForGetOrHead(key string, additionalQuery map[string]string) string {
	query := make(url.Values)
	for k, v := range additionalQuery {
		query.Add(k, v)
	}
	queryStr := query.Encode()
	if queryStr != "" {
		queryStr = "&" + queryStr
	}
	var domain string
	if output.isCustomDomain {
		domain = output.host
	} else {
		domain = fmt.Sprintf("%s.%s", output.bucket, output.host)
	}
	var str string
	if output.SignatureQuery != "" {
		str = fmt.Sprintf("%s://%s/%s?%s%s", output.scheme, domain, key, output.SignatureQuery, queryStr)
	} else {
		str = fmt.Sprintf("%s://%s/%s", output.scheme, domain, key)
	}
	return str
}

type PreSignedURLInput struct {
	HTTPMethod          enum.HttpMethodType
	Bucket              string
	Key                 string
	Expires             int64 // Expiration time in seconds, default 3600 seconds, max 7 days, range [1, 604800]
	Header              map[string]string
	Query               map[string]string
	AlternativeEndpoint string
	IsCustomDomain      *bool
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
	ProjectName      string                `location:"header" locationName:"X-Tos-Project-Name"`
	BucketType       enum.BucketType       `location:"header" locationName:"X-Tos-Bucket-Type"`
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
	ProjectName  string                `json:"ProjectName"`
	BucketType   enum.BucketType       `json:"BucketType"`
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
	ResponseVary  bool     `json:"ResponseVary,omitempty"`
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
	CreationDate     string          `json:"CreationDate,omitempty"`
	Name             string          `json:"Name,omitempty"`
	Location         string          `json:"Location,omitempty"`
	ExtranetEndpoint string          `json:"ExtranetEndpoint,omitempty"`
	IntranetEndpoint string          `json:"IntranetEndpoint,omitempty"`
	ProjectName      string          `json:"ProjectName,omitempty"`
	BucketType       enum.BucketType `json:"BucketType,omitempty"`
}

type ListBucketsInput struct {
	ProjectName string          `location:"header" locationName:"X-Tos-Project-Name"`
	BucketType  enum.BucketType `location:"header" locationName:"X-Tos-Bucket-Type"`
}

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

	Callback                  string                `location:"header" locationName:"X-Tos-Callback"`
	CallbackVar               string                `location:"header" locationName:"X-Tos-Callback-Var"`
	WebsiteRedirectLocation   string                `location:"header" locationName:"X-Tos-Website-Redirect-Location"`
	StorageClass              enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
	SSECAlgorithm             string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey                   string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5                string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption      string                `location:"header" locationName:"X-Tos-Server-Side-Encryption"`
	ServerSideEncryptionKeyID string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Kms-Key-Id"`
	TrafficLimit              int64                 `location:"header" locationName:"X-Tos-Traffic-Limit"`
	ForbidOverwrite           bool                  `location:"header" locationName:"X-Tos-Forbid-Overwrite"`
	IfMatch                   string                `location:"header" locationName:"X-Tos-If-Match"`
	Tagging                   string                `location:"header" locationName:"X-Tos-Tagging"`
	ObjectExpires             int64                 `location:"header" locationName:"X-Tos-Object-Expires"`
	Meta                      map[string]string     `location:"headers"`
	DataTransferListener      DataTransferListener
	RateLimiter               RateLimiter
}

type PutObjectV2Input struct {
	PutObjectBasicInput
	Content io.Reader
}

type PutObjectV2Output struct {
	RequestInfo
	ETag                      string
	SSECAlgorithm             string
	SSECKeyMD5                string
	VersionID                 string
	CallbackResult            string
	HashCrc64ecma             uint64
	ServerSideEncryption      string
	ServerSideEncryptionKeyID string
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
	TrafficLimit            int64                 `location:"header" locationName:"X-Tos-Traffic-Limit"`
	IfMatch                 string                `location:"header" locationName:"X-Tos-If-Match"`

	ObjectExpires        int64             `location:"header" locationName:"X-Tos-Object-Expires"`
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
	VersionID string `location:"query" locationName:"versionId"`

	CacheControl       string    `location:"header" locationName:"Cache-Control"`
	ContentDisposition string    `location:"header" locationName:"Content-Disposition"`
	ContentEncoding    string    `location:"header" locationName:"Content-Encoding"`
	ContentLanguage    string    `location:"header" locationName:"Content-Language"`
	ContentType        string    `location:"header" locationName:"Content-Type"`
	Expires            time.Time `location:"header" locationName:"Expires"`
	ObjectExpires      int64     `location:"header" locationName:"X-Tos-Object-Expires"`

	Meta map[string]string `location:"headers"`
}

type SetObjectMetaOutput struct {
	RequestInfo `json:"-"`
}

type ListObjectsV2Input struct {
	Bucket string
	ListObjectsInput
}

type ListObjectsType2Input struct {
	Bucket            string
	Prefix            string `location:"query" locationName:"prefix"`
	Delimiter         string `location:"query" locationName:"delimiter"`
	StartAfter        string `location:"query" locationName:"start-after"`
	ContinuationToken string `location:"query" locationName:"continuation-token"`
	MaxKeys           int    `location:"query" locationName:"max-keys"`
	EncodingType      string `location:"query" locationName:"encoding-type"`
	FetchMeta         bool   `location:"query" locationName:"fetch-meta"`
	ListOnlyOnce      bool
}

type ListObjectsType2Output struct {
	RequestInfo
	Name                  string               `json:"Name,omitempty"`
	Prefix                string               `json:"Prefix,omitempty"`
	ContinuationToken     string               `json:"ContinuationToken,omitempty"`
	KeyCount              int                  `json:"KeyCount,omitempty"`
	MaxKeys               int                  `json:"MaxKeys,omitempty"`
	Delimiter             string               `json:"Delimiter,omitempty"`
	IsTruncated           bool                 `json:"IsTruncated,omitempty"`
	EncodingType          string               `json:"EncodingType,omitempty"`
	NextContinuationToken string               `json:"NextContinuationToken,omitempty"`
	CommonPrefixes        []ListedCommonPrefix `json:"CommonPrefixes,omitempty"`
	Contents              []ListedObjectV2     `json:"Contents,omitempty"`
}

type ListObjectsInput struct {
	Prefix       string `location:"query" locationName:"prefix"`
	Delimiter    string `location:"query" locationName:"delimiter"`
	Marker       string `location:"query" locationName:"marker"`
	MaxKeys      int    `location:"query" locationName:"max-keys"`
	EncodingType string `location:"query" locationName:"encoding-type"` // "" or "url"
	FetchMeta    bool   `location:"query" locationName:"fetch-meta"`
	// Deprecated
	Reverse bool
}

type ListedObject struct {
	Key           string   `json:"Key,omitempty"`
	LastModified  string   `json:"LastModified,omitempty"`
	ETag          string   `json:"ETag,omitempty"`
	Size          int64    `json:"Size,omitempty"`
	Owner         Owner    `json:"Owner,omitempty"`
	StorageClass  string   `json:"StorageClass,omitempty"`
	Type          string   `json:"Type,omitempty"`
	Meta          Metadata `json:"UserMeta,omitempty"`
	HashCrc64ecma uint64   `json:"HashCrc64Ecma,omitempty"`
	// Deprecated: ues Type instead
	ObjectType string `json:"-"`
}

type listedObject struct {
	Key           string     `json:"Key,omitempty"`
	LastModified  string     `json:"LastModified,omitempty"`
	ETag          string     `json:"ETag,omitempty"`
	Size          int64      `json:"Size,omitempty"`
	Owner         Owner      `json:"Owner,omitempty"`
	StorageClass  string     `json:"StorageClass,omitempty"`
	Type          string     `json:"Type,omitempty"`
	Meta          []userMeta `json:"UserMeta,omitempty"`
	HashCrc64ecma string     `json:"HashCrc64Ecma,omitempty"`
	// Deprecated: ues Type instead
	ObjectType string `json:"-"`
}

type ListedObjectV2 struct {
	Key           string
	LastModified  time.Time
	ETag          string
	Size          int64
	Owner         Owner
	StorageClass  enum.StorageClassType
	HashCrc64ecma uint64
	Meta          Metadata
	ObjectType    string `json:"Type,omitempty"`
}

type userMeta struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type listedObjectV2 struct {
	Key           string                `json:"Key,omitempty"`
	LastModified  time.Time             `json:"LastModified,omitempty"`
	ETag          string                `json:"ETag,omitempty"`
	Size          int64                 `json:"Size,omitempty"`
	Owner         Owner                 `json:"Owner,omitempty"`
	StorageClass  enum.StorageClassType `json:"StorageClass,omitempty"`
	HashCrc64ecma string                `json:"HashCrc64Ecma,omitempty"`
	Meta          []userMeta            `json:"UserMeta,omitempty"`
	ObjectType    string                `json:"Type,omitempty"`
}

type ListedCommonPrefix struct {
	Prefix       string     `json:"Prefix,omitempty"`
	LastModified *time.Time `json:"LastModified,omitempty"`
}

type listObjectsOutput struct {
	Name           string               `json:"Name,omitempty"` // bucket name
	Prefix         string               `json:"Prefix,omitempty"`
	Marker         string               `json:"Marker,omitempty"`
	MaxKeys        int64                `json:"MaxKeys,omitempty"`
	NextMarker     string               `json:"NextMarker,omitempty"`
	Delimiter      string               `json:"Delimiter,omitempty"`
	IsTruncated    bool                 `json:"IsTruncated,omitempty"`
	EncodingType   string               `json:"EncodingType,omitempty"`
	CommonPrefixes []ListedCommonPrefix `json:"CommonPrefixes,omitempty"`
	Contents       []listedObject       `json:"Contents,omitempty"`
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

type listObjectsType2Output struct {
	RequestInfo           `json:"-"`
	Name                  string               `json:"Name,omitempty"`
	Prefix                string               `json:"Prefix,omitempty"`
	ContinuationToken     string               `json:"ContinuationToken,omitempty"`
	KeyCount              int                  `json:"KeyCount,omitempty"`
	MaxKeys               int                  `json:"MaxKeys,omitempty"`
	Delimiter             string               `json:"Delimiter,omitempty"`
	IsTruncated           bool                 `json:"IsTruncated,omitempty"`
	EncodingType          string               `json:"EncodingType,omitempty"`
	NextContinuationToken string               `json:"NextContinuationToken,omitempty"`
	CommonPrefixes        []ListedCommonPrefix `json:"CommonPrefixes,omitempty"`
	Contents              []listedObjectV2     `json:"Contents,omitempty"`
}

type ListObjectVersionsInput struct {
	Prefix          string `location:"query" locationName:"prefix"`
	Delimiter       string `location:"query" locationName:"delimiter"`
	KeyMarker       string `location:"query" locationName:"key-marker"`
	VersionIDMarker string `location:"query" locationName:"version-id-marker"`
	MaxKeys         int    `location:"query" locationName:"max-keys"`
	EncodingType    string `location:"query" locationName:"encoding-type"` // "" or "url"
	FetchMeta       bool   `location:"query" locationName:"fetch-meta"`
}

type ListObjectVersionsV2Input struct {
	Bucket string `json:"Prefix,omitempty"`
	ListObjectVersionsInput
}

type ListedObjectVersion struct {
	ETag         string   `json:"ETag,omitempty"`
	IsLatest     bool     `json:"IsLatest,omitempty"`
	Key          string   `json:"Key,omitempty"`
	LastModified string   `json:"LastModified,omitempty"`
	Owner        Owner    `json:"Owner,omitempty"`
	Size         int64    `json:"Size,omitempty"`
	StorageClass string   `json:"StorageClass,omitempty"`
	Type         string   `json:"Type,omitempty"`
	VersionID    string   `json:"VersionId,omitempty"`
	Meta         Metadata `json:"UserMeta,omitempty"`

	HashCrc64ecma uint64 `json:"HashCrc64Ecma,omitempty"`
	ObjectType    string `json:"-"`
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
	Meta          []userMeta `json:"UserMeta,omitempty"`
	ObjectType    string     `json:"Type,omitempty"`
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
	Meta          Metadata `json:"UserMeta,omitempty"`
	ObjectType    string   `json:"Type,omitempty"`
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

type listedObjectVersion struct {
	ETag          string     `json:"ETag,omitempty"`
	IsLatest      bool       `json:"IsLatest,omitempty"`
	Key           string     `json:"Key,omitempty"`
	LastModified  string     `json:"LastModified,omitempty"`
	Owner         Owner      `json:"Owner,omitempty"`
	Size          int64      `json:"Size,omitempty"`
	StorageClass  string     `json:"StorageClass,omitempty"`
	Type          string     `json:"Type,omitempty"`
	VersionID     string     `json:"VersionId,omitempty"`
	Meta          []userMeta `json:"UserMeta,omitempty"`
	HashCrc64ecma string     `json:"HashCrc64Ecma,omitempty"`
	// Deprecated: ues Type instead
	ObjectType string `json:"-"`
}

type listObjectVersionsOutput struct {
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
	Versions            []listedObjectVersion     `json:"Versions,omitempty"`
	DeleteMarkers       []ListedDeleteMarkerEntry `json:"DeleteMarkers,omitempty"`
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
	TrafficLimit  int64  `location:"header" locationName:"X-Tos-Traffic-Limit"`

	ResponseCacheControl       string                 `location:"query" locationName:"response-cache-control"`
	ResponseContentDisposition string                 `location:"query" locationName:"response-content-disposition"`
	ResponseContentEncoding    string                 `location:"query" locationName:"response-content-encoding"`
	ResponseContentLanguage    string                 `location:"query" locationName:"response-content-language"`
	ResponseContentType        string                 `location:"query" locationName:"response-content-type"`
	ResponseExpires            time.Time              `location:"query" locationName:"response-expires"`
	Process                    string                 `location:"query" locationName:"x-tos-process"`
	SaveBucket                 string                 `location:"query" locationName:"x-tos-save-bucket"`
	SaveObject                 string                 `location:"query" locationName:"x-tos-save-object"`
	DocPage                    int                    `location:"query" locationName:"x-tos-doc-page"`
	SrcType                    enum.DocPreviewSrcType `location:"query" locationName:"x-tos-doc-src-type"`
	DstType                    enum.DocPreviewDstType `location:"query" locationName:"x-tos-doc-dst-type"`

	RangeStart int64
	RangeEnd   int64
	Range      string

	DataTransferListener DataTransferListener
	RateLimiter          RateLimiter
	// Deprecated Not Use
	PartNumber int
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
	SymlinkTargetSize int64 `location:"header" locationName:"X-Tos-Symlink-Target-Size"`
}

type HeadObjectV2Output struct {
	RequestInfo `json:"-"`
	ObjectMetaV2
	SymlinkTargetSize int64 `location:"header" locationName:"X-Tos-Symlink-Target-Size"`
}

type DeleteObjectV2Input struct {
	Bucket    string
	Key       string
	VersionID string `location:"query" locationName:"versionId"`
	Recursive bool
}

type DeleteObjectOutput struct {
	RequestInfo  `json:"-"`
	DeleteMarker bool   `json:"DeleteMarker,omitempty"`
	VersionID    string `json:"VersionId,omitempty"`
	TrashPath    string `json:"TrashPath,omitempty"`
}

type DeleteObjectV2Output struct {
	DeleteObjectOutput
}

type ObjectTobeDeleted struct {
	Key       string `json:"Key,omitempty"`
	VersionID string `json:"VersionId,omitempty"`
}

type DeleteMultiObjectsInput struct {
	Bucket    string
	Objects   []ObjectTobeDeleted `json:"Objects,omitempty"`
	Quiet     bool                `json:"Quiet,omitempty"`
	Recursive bool                `json:"-"`
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
	TrashPath             string `json:"TrashPath,omitempty"`
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
	SrcVersionID       string       `location:"query" locationName:"versionId"`
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

	CopySourceSSECAlgorithm   string `location:"header" locationName:"X-Tos-Copy-Source-Server-Side-Encryption-Customer-Algorithm"`
	CopySourceSSECKey         string `location:"header" locationName:"X-Tos-Copy-Source-Server-Side-Encryption-Customer-Key"`
	CopySourceSSECKeyMD5      string `location:"header" locationName:"X-Tos-Copy-Source-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption      string `location:"header" locationName:"X-Tos-Server-Side-Encryption"`
	ServerSideEncryptionKeyID string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Kms-Key-Id"`

	SSECKey           string                     `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5        string                     `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	SSECAlgorithm     string                     `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	TrafficLimit      int64                      `location:"header" locationName:"X-Tos-Traffic-Limit"`
	ForbidOverwrite   bool                       `location:"header" locationName:"X-Tos-Forbid-Overwrite"`
	IfMatch           string                     `location:"header" locationName:"X-Tos-If-Match"`
	MetadataDirective enum.MetadataDirectiveType `location:"header" locationName:"X-Tos-Metadata-Directive"`
	Tagging           string                     `location:"header" locationName:"X-Tos-Tagging"`
	ObjectExpires     int64                      `location:"header" locationName:"X-Tos-Object-Expires"`
	TaggingDirective  enum.TaggingDirectiveType  `location:"header" locationName:"X-Tos-Tagging-Directive"`
	Meta              map[string]string          `location:"headers"`
}

type copyObjectOutput struct {
	ETag         string `json:"ETag,omitempty"`         // at body
	LastModified string `json:"LastModified,omitempty"` // at body
	Error
}

type CopyObjectOutput struct {
	RequestInfo               `json:"-"`
	VersionID                 string `json:"VersionId,omitempty"`
	SourceVersionID           string `json:"SourceVersionId,omitempty"`
	ETag                      string `json:"ETag,omitempty"`         // at body
	LastModified              string `json:"LastModified,omitempty"` // at body
	SSECAlgorithm             string `json:"SSECAlgorithm,omitempty"`
	SSECKeyMD5                string `json:"SSECKeyMD5,omitempty"`
	ServerSideEncryption      string `json:"ServerSideEncryption,omitempty"`
	ServerSideEncryptionKeyID string `json:"ServerSideEncryptionKmsKeyId,omitempty"`
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
	CopySourceRange      string

	CopySourceIfMatch           string    `location:"header" locationName:"X-Tos-Copy-Source-If-Match"`
	CopySourceIfModifiedSince   time.Time `location:"header" locationName:"X-Tos-Copy-Source-If-Modified-Since"`
	CopySourceIfNoneMatch       string    `location:"header" locationName:"X-Tos-Copy-Source-If-None-Match"`
	CopySourceIfUnmodifiedSince time.Time `location:"header" locationName:"X-Tos-Copy-Source-If-Unmodified-Since"`

	CopySourceSSECAlgorithm string `location:"header" locationName:"X-Tos-Copy-Source-Server-Side-Encryption-Customer-Algorithm"`
	CopySourceSSECKey       string `location:"header" locationName:"X-Tos-Copy-Source-Server-Side-Encryption-Customer-Key"`
	CopySourceSSECKeyMD5    string `location:"header" locationName:"X-Tos-Copy-Source-Server-Side-Encryption-Customer-Key-MD5"`

	SSECKey       string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5    string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	SSECAlgorithm string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	TrafficLimit  int64  `location:"header" locationName:"X-Tos-Traffic-Limit"`
}

type UploadPartCopyV2Output struct {
	RequestInfo
	PartNumber                int
	ETag                      string
	LastModified              time.Time
	CopySourceVersionID       string
	ServerSideEncryption      string
	ServerSideEncryptionKeyID string
	SSECAlgorithm             string
	SSECKeyMD5                string
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

	WebsiteRedirectLocation   string                `location:"header" locationName:"X-Tos-Website-Redirect-Location"`
	StorageClass              enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
	SSECAlgorithm             string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey                   string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5                string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption      string                `location:"header" locationName:"X-Tos-Server-Side-Encryption"`
	ServerSideEncryptionKeyID string                `location:"header" locationName:"X-Tos-Server-Side-Encryption-Kms-Key-Id"`
	ForbidOverwrite           bool                  `location:"header" locationName:"X-Tos-Forbid-Overwrite"`
	Tagging                   string                `location:"header" locationName:"X-Tos-Tagging"`
	ObjectExpires             int64                 `location:"header" locationName:"X-Tos-Object-Expires"`
	Meta                      map[string]string     `location:"headers"`
}

type RenameObjectInput struct {
	Bucket         string
	Key            string
	NewKey         string `location:"query" locationName:"name"`
	RecursiveMkdir bool   `location:"header" locationName:"X-Tos-Recursive-Mkdir"`
}

type RenameObjectOutput struct {
	RequestInfo
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
	RequestInfo               `json:"-"`
	Bucket                    string `json:"Bucket,omitempty"`
	Key                       string `json:"Key,omitempty"`
	UploadID                  string `json:"UploadID,omitempty"`
	SSECAlgorithm             string `json:"SSECAlgorithm,omitempty"`
	SSECKeyMD5                string `json:"SSECKeyMD5,omitempty"`
	EncodingType              string `json:"EncodingType,omitempty"`
	ServerSideEncryption      string `json:"ServerSideEncryption,omitempty"`
	ServerSideEncryptionKeyID string `json:"ServerSideEncryptionKeyID,omitempty"`
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

	SSECAlgorithm             string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Algorithm"`
	SSECKey                   string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key"`
	SSECKeyMD5                string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Customer-Key-MD5"`
	ServerSideEncryption      string `location:"header" locationName:"X-Tos-Server-Side-Encryption"`
	ServerSideEncryptionKeyID string `location:"header" locationName:"X-Tos-Server-Side-Encryption-Kms-Key-Id"`

	TrafficLimit int64 `location:"header" locationName:"X-Tos-Traffic-Limit"`

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
	PartNumber                int
	ETag                      string
	SSECAlgorithm             string
	SSECKeyMD5                string
	HashCrc64ecma             uint64
	ServerSideEncryption      string
	ServerSideEncryptionKeyID string
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
	RequestInfo               `json:"-"`
	VersionID                 string `json:"VersionId,omitempty"`
	Bucket                    string
	Key                       string
	ETag                      string
	Location                  string
	ServerSideEncryption      string
	ServerSideEncryptionKeyID string
}

type CompleteMultipartUploadV2Input struct {
	Bucket          string
	Key             string
	CompleteAll     bool
	UploadID        string `location:"query" locationName:"uploadId"`
	Callback        string `location:"header" locationName:"X-Tos-Callback"`
	CallbackVar     string `location:"header" locationName:"X-Tos-Callback-Var"`
	ForbidOverwrite bool   `location:"header" locationName:"X-Tos-Forbid-Overwrite"`
	Parts           []UploadedPartV2
}

type CompleteMultipartUploadV2Output struct {
	RequestInfo
	Bucket                    string
	Key                       string
	ETag                      string
	Location                  string
	CompletedParts            []UploadedPartV2
	VersionID                 string
	HashCrc64ecma             uint64
	CallbackResult            string
	ServerSideEncryption      string
	ServerSideEncryptionKeyID string
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
	Prefix         string `location:"query" locationName:"prefix"`
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

type putBucketLifecycleInput struct {
	Rules []lifecycleRule `json:"Rules,omitempty"`
}

type lifecycleRule struct {
	ID                             string                          `json:"ID,omitempty"`
	Prefix                         string                          `json:"Prefix,omitempty"`
	Status                         enum.StatusType                 `json:"Status,omitempty"`
	Transitions                    []transition                    `json:"Transitions,omitempty"`
	Expiration                     *expiration                     `json:"Expiration,omitempty"`
	NonCurrentVersionTransition    []NonCurrentVersionTransition   `json:"NoncurrentVersionTransitions,omitempty"`
	NoCurrentVersionExpiration     *NoCurrentVersionExpiration     `json:"NoncurrentVersionExpiration,omitempty"`
	Tag                            []Tag                           `json:"Tags,omitempty"`
	AbortInCompleteMultipartUpload *AbortInCompleteMultipartUpload `json:"AbortIncompleteMultipartUpload,omitempty"`
	Filter                         *LifecycleRuleFilter            `json:"Filter,omitempty"`
}

type PutBucketLifecycleInput struct {
	Bucket                 string
	Rules                  []LifecycleRule `json:"Rules,omitempty"`
	AllowSameActionOverlap bool            `location:"header" locationName:"X-Tos-Allow-Same-Action-Overlap"`
}

type GetBucketLifecycleInput struct {
	Bucket string
}

type GetBucketLifecycleOutput struct {
	RequestInfo
	Rules []LifecycleRule `json:"Rules"`
}

type DeleteBucketLifecycleInput struct {
	Bucket string
}

type DeleteBucketLifecycleOutput struct {
	RequestInfo
}

type LifecycleRule struct {
	ID                             string                          `json:"ID,omitempty"`
	Prefix                         string                          `json:"Prefix,omitempty"`
	Status                         enum.StatusType                 `json:"Status,omitempty"`
	Transitions                    []Transition                    `json:"Transitions,omitempty"`
	Expiration                     *Expiration                     `json:"Expiration,omitempty"`
	NonCurrentVersionTransition    []NonCurrentVersionTransition   `json:"NoncurrentVersionTransitions,omitempty"`
	NoCurrentVersionExpiration     *NoCurrentVersionExpiration     `json:"NoncurrentVersionExpiration,omitempty"`
	Tag                            []Tag                           `json:"Tags,omitempty"`
	AbortInCompleteMultipartUpload *AbortInCompleteMultipartUpload `json:"AbortIncompleteMultipartUpload,omitempty"`
	Filter                         *LifecycleRuleFilter            `json:"Filter,omitempty"`
}

type LifecycleRuleFilter struct {
	ObjectSizeGreaterThan   int             `json:"ObjectSizeGreaterThan"`
	GreaterThanIncludeEqual enum.StatusType `json:"GreaterThanIncludeEqual"`
	ObjectSizeLessThan      int             `json:"ObjectSizeLessThan"`
	LessThanIncludeEqual    enum.StatusType `json:"LessThanIncludeEqual"`
}

type AbortInCompleteMultipartUpload struct {
	DaysAfterInitiation int `json:"DaysAfterInitiation,omitempty"`
}

type Tag struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type NoCurrentVersionExpiration struct {
	NoCurrentDays  int        `json:"NoncurrentDays,omitempty"`
	NonCurrentDate *time.Time `json:"NoncurrentDate,omitempty"`
}

type NonCurrentVersionTransition struct {
	NonCurrentDays int                   `json:"NoncurrentDays,omitempty"`
	NonCurrentDate *time.Time            `json:"NoncurrentDate,omitempty"`
	StorageClass   enum.StorageClassType `json:"StorageClass,omitempty"`
}

type transition struct {
	Days         int                   `json:"Days,omitempty"`
	Date         string                `json:"Date,omitempty"`
	StorageClass enum.StorageClassType `json:"StorageClass,omitempty"`
}

type Transition struct {
	Days         int                   `json:"Days,omitempty"`
	Date         time.Time             `json:"Date,omitempty"`
	StorageClass enum.StorageClassType `json:"StorageClass,omitempty"`
}

type Expiration struct {
	Days int       `json:"Days,omitempty"`
	Date time.Time `json:"Date,omitempty"`
}

type expiration struct {
	Days int    `json:"Days,omitempty"`
	Date string `json:"Date,omitempty"`
}

type PutLifecycleOutput struct {
	RequestInfo
}

type PutBucketMirrorBackOutput struct {
	RequestInfo
}

type putBucketMirrorBackInput struct {
	Rules []MirrorBackRule `json:"Rules"`
}

type PutBucketMirrorBackInput struct {
	Bucket string
	Rules  []MirrorBackRule
}

type MirrorBackRule struct {
	ID        string    `json:"ID,omitempty"`
	Condition Condition `json:"Condition,omitempty"`
	Redirect  Redirect  `json:"Redirect,omitempty"`
}

type Condition struct {
	HttpCode   int      `json:"HttpCode,omitempty"`
	KeyPrefix  string   `json:"KeyPrefix,omitempty"`
	KeySuffix  string   `json:"KeySuffix,omitempty"`
	AllowHost  []string `xml:"AllowHost" json:"AllowHost,omitempty"`
	HttpMethod []string `json:"HttpMethod,omitempty"`
}

type Redirect struct {
	RedirectType                   enum.RedirectType           `json:"RedirectType,omitempty"`
	FetchSourceOnRedirect          bool                        `json:"FetchSourceOnRedirect,omitempty"`
	PassQuery                      bool                        `json:"PassQuery,omitempty"`
	FollowRedirect                 bool                        `json:"FollowRedirect,omitempty"`
	MirrorHeader                   MirrorHeader                `json:"MirrorHeader,omitempty"`
	PublicSource                   PublicSource                `json:"PublicSource,omitempty"`
	Transform                      Transform                   `json:"Transform,omitempty"`
	FetchHeaderToMetaDataRules     []FetchHeaderToMetaDataRule `json:"FetchHeaderToMetaDataRules,omitempty"`
	PrivateSource                  *PrivateSource              `json:"PrivateSource,omitempty"`
	FetchSourceOnRedirectWithQuery *bool                       `json:"FetchSourceOnRedirectWithQuery,omitempty"`
}
type PrivateSource struct {
	SourceEndpoint CommonSourceEndpoint
}

type CommonSourceEndpoint struct {
	Primary  []EndpointCredentialProvider `json:"Primary,omitempty"`
	Follower []EndpointCredentialProvider `json:"Follower,omitempty"`
}

type EndpointCredentialProvider struct {
	Endpoint           string              `json:"Endpoint,omitempty"`
	BucketName         string              `json:"BucketName,omitempty"`
	CredentialProvider *CredentialProvider `json:"CredentialProvider,omitempty"`
}

type CredentialProvider struct {
	Role string `json:"Role,omitempty"`
}

type FetchHeaderToMetaDataRule struct {
	SourceHeader   string `json:"SourceHeader"`
	MetaDataSuffix string `json:"MetaDataSuffix"`
}

type Transform struct {
	WithKeyPrefix    string           `json:"WithKeyPrefix,omitempty"`
	WithKeySuffix    string           `json:"WithKeySuffix,omitempty"`
	ReplaceKeyPrefix ReplaceKeyPrefix `json:"ReplaceKeyPrefix,omitempty"`
}

type ReplaceKeyPrefix struct {
	KeyPrefix   string `json:"KeyPrefix,omitempty"`
	ReplaceWith string `json:"ReplaceWith,omitempty"`
}

type PublicSource struct {
	SourceEndpoint SourceEndpoint `json:"SourceEndpoint,omitempty"`
	FixedEndpoint  bool           `json:"FixedEndpoint,omitempty"`
}

type GetBucketMirrorBackInput struct {
	Bucket string
}

type GetBucketMirrorBackOutput struct {
	RequestInfo
	Rules []MirrorBackRule
}

type DeleteObjectTaggingInput struct {
	Bucket    string
	Key       string
	VersionID string `location:"query" locationName:"versionId"`
}

type GetObjectTaggingInput struct {
	Bucket    string
	Key       string
	VersionID string `location:"query" locationName:"versionId"`
}
type putObjectTaggingInput struct {
	TagSet TagSet `json:"TagSet"`
}
type PutObjectTaggingInput struct {
	Bucket    string
	Key       string
	VersionID string `location:"query" locationName:"versionId"`
	TagSet    TagSet `json:"TagSet"`
}

type PutObjectTaggingOutput struct {
	RequestInfo
	VersionID string
}

type GetObjectTaggingOutput struct {
	RequestInfo
	VersionID string
	TagSet    TagSet
}

type TagSet struct {
	Tags []Tag
}

type DeleteObjectTaggingOutput struct {
	RequestInfo
	VersionID string
}

type DeleteBucketMirrorBackInput struct {
	Bucket string
}

type DeleteBucketMirrorBackOutput struct {
	RequestInfo
}

type MirrorHeaderKeyValue struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type SourceEndpoint struct {
	Primary  []string `json:"Primary,omitempty"`
	Follower []string `json:"Follower,omitempty"`
}
type MirrorHeader struct {
	PassAll bool                   `json:"PassAll,omitempty"`
	Pass    []string               `json:"Pass,omitempty"`
	Remove  []string               `json:"Remove,omitempty"`
	Set     []MirrorHeaderKeyValue `json:"Set,omitempty"`
}

type PutBucketStorageClassInput struct {
	Bucket       string
	StorageClass enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
}

type PutBucketStorageClassOutput struct {
	RequestInfo
}

type GetBucketLocationInput struct {
	Bucket string
}

type GetBucketLocationOutput struct {
	RequestInfo      `json:"-"`
	Region           string `json:"Region,omitempty"`
	ExtranetEndpoint string `json:"ExtranetEndpoint,omitempty"`
	IntranetEndpoint string `json:"IntranetEndpoint,omitempty"`
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
	filePath              string
	PartSize              int64
	TaskNum               int
	EnableCheckpoint      bool
	CheckpointFile        string
	tempFile              string
	TrafficLimit          int64
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

	FilePath         string
	PartSize         int64
	TaskNum          int
	EnableCheckpoint bool
	CheckpointFile   string
	TrafficLimit     int64 `location:"header" locationName:"X-Tos-Traffic-Limit"`

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
	RetryCount    int // The DownloadFile and UploadFile method do not support this field
}

type putBucketNotificationInput struct {
	CloudFunctionConfigurations []CloudFunctionConfiguration `json:"CloudFunctionConfigurations"`
	RocketMQConfigurations      []RocketMQConfiguration      `json:"RocketMQConfigurations"`
}

type RocketMQConf struct {
	InstanceID  string `json:"InstanceId"`
	Topic       string `json:"Topic"`
	AccessKeyID string `json:"AccessKeyId"`
}

type RocketMQConfiguration struct {
	ID       string       `json:"RuleId"`
	Role     string       `json:"Role"`
	Events   []string     `json:"Events"`
	Filter   Filter       `json:"Filter"`
	RocketMQ RocketMQConf `json:"RocketMQ"`
}

type PutBucketNotificationInput struct {
	Bucket                      string                       `json:"-"`
	CloudFunctionConfigurations []CloudFunctionConfiguration `json:"CloudFunctionConfigurations"`
	RocketMQConfigurations      []RocketMQConfiguration      `json:"RocketMQConfigurations"`
}

type PutBucketNotificationOutput struct {
	RequestInfo
}

type CloudFunctionConfiguration struct {
	ID            string   `json:"RuleId"`
	Events        []string `json:"Events"`
	Filter        Filter   `json:"Filter"`
	CloudFunction string   `json:"CloudFunction"`
}

type Filter struct {
	Key FilterKey `json:"TOSKey"`
}

type FilterKey struct {
	Rules []FilterRule `json:"FilterRules"`
}

type FilterRule struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type GetBucketNotificationInput struct {
	Bucket string
}

type GetBucketNotificationOutput struct {
	RequestInfo
	CloudFunctionConfigurations []CloudFunctionConfiguration `json:"CloudFunctionConfigurations"`
	RocketMQConfigurations      []RocketMQConfiguration      `json:"RocketMQConfigurations"`
}

type GetBucketNotificationType2Input struct {
	Bucket string
}

type NotificationRule struct {
	RuleID      string                  `json:"RuleId"`
	Events      []string                `json:"Events"` // 支持的值在不断增加，不定义成枚举
	Filter      NotificationFilter      `json:"Filter"`
	Destination NotificationDestination `json:"Destination"`
}

type NotificationFilter struct {
	TOSKey NotificationFilterKey `json:"TOSKey"`
}

type NotificationFilterKey struct {
	FilterRules []NotificationFilterRule `json:"FilterRules"`
}

type NotificationFilterRule struct {
	Name  string `json:"Name"`
	Value string `json:"Value"`
}

type NotificationDestination struct {
	RocketMQ []DestinationRocketMQ `json:"RocketMQ"`
	VeFaaS   []DestinationVeFaaS   `json:"VeFaaS"`
}

type DestinationRocketMQ struct {
	Role        string `json:"Role"`
	InstanceID  string `json:"InstanceId"`
	Topic       string `json:"Topic"`
	AccessKeyID string `json:"AccessKeyId"`
}

type DestinationVeFaaS struct {
	FunctionID string `json:"FunctionId"`
}

type GetBucketNotificationType2Output struct {
	RequestInfo
	Rules   []NotificationRule `json:"Rules"`
	Version string             `json:"Version"`
}

type putBucketNotificationType2Input struct {
	Rules   []NotificationRule `json:"Rules"`
	Version string             `json:"Version"`
}

type PutBucketNotificationType2Input struct {
	Bucket  string
	Rules   []NotificationRule `json:"Rules"`
	Version string             `json:"Version"`
}

type PutBucketNotificationType2Output struct {
	RequestInfo
}

type putBucketVersioningInput struct {
	Status enum.VersioningStatusType `json:"Status"`
}

type PutBucketVersioningInput struct {
	Bucket string
	Status enum.VersioningStatusType
}

type PutBucketVersioningOutput struct {
	RequestInfo
}

type GetBucketVersioningInput struct {
	Bucket string
}

type GetBucketVersioningOutputV2 struct {
	RequestInfo
	Status enum.VersioningStatusType `json:"Status"`
}

type putBucketWebsiteInput struct {
	RedirectAllRequestsTo *RedirectAllRequestsTo `json:"RedirectAllRequestsTo,omitempty"`
	IndexDocument         *IndexDocument         `json:"IndexDocument,omitempty"`
	ErrorDocument         *ErrorDocument         `json:"ErrorDocument,omitempty"`
	RoutingRules          []RoutingRule          `json:"RoutingRules,omitempty"`
}

type PutBucketWebsiteInput struct {
	Bucket                string
	RedirectAllRequestsTo *RedirectAllRequestsTo `json:"RedirectAllRequestsTo,omitempty"`
	IndexDocument         *IndexDocument         `json:"IndexDocument,omitempty"`
	ErrorDocument         *ErrorDocument         `json:"ErrorDocument,omitempty"`
	RoutingRules          *RoutingRules          `json:"RoutingRules,omitempty"`
}

type RedirectAllRequestsTo struct {
	HostName string `json:"HostName"`
	Protocol string `json:"Protocol,omitempty"`
}

type IndexDocument struct {
	Suffix          string `json:"Suffix"`
	ForbiddenSubDir bool   `json:"ForbiddenSubDir,omitempty"`
}

type ErrorDocument struct {
	Key string `json:"Key"`
}

type RoutingRules struct {
	Rules []RoutingRule `json:"RoutingRules,omitempty"`
}

type RoutingRule struct {
	Condition RoutingRuleCondition `json:"Condition"`
	Redirect  RoutingRuleRedirect  `json:"Redirect"`
}

type RoutingRuleCondition struct {
	KeyPrefixEquals             string `json:"KeyPrefixEquals,omitempty"`
	HttpErrorCodeReturnedEquals int    `json:"HttpErrorCodeReturnedEquals,omitempty"`
}

type RoutingRuleRedirect struct {
	Protocol             enum.ProtocolType `json:"Protocol,omitempty"`
	HostName             string            `json:"HostName,omitempty"`
	ReplaceKeyPrefixWith string            `json:"ReplaceKeyPrefixWith,omitempty"`
	ReplaceKeyWith       string            `json:"ReplaceKeyWith,omitempty"`
	HttpRedirectCode     int               `json:"HttpRedirectCode,omitempty"`
}

type PutBucketWebsiteOutput struct {
	RequestInfo
}

type GetBucketWebsiteInput struct {
	Bucket string
}

type GetBucketWebsiteOutput struct {
	RequestInfo
	RedirectAllRequestsTo *RedirectAllRequestsTo `json:"RedirectAllRequestsTo,omitempty"`
	IndexDocument         *IndexDocument         `json:"IndexDocument,omitempty"`
	ErrorDocument         *ErrorDocument         `json:"ErrorDocument,omitempty"`
	RoutingRules          []RoutingRule          `json:"RoutingRules,omitempty"`
}

type DeleteBucketWebsiteInput struct {
	Bucket string
}

type DeleteBucketWebsiteOutput struct {
	RequestInfo
}

type putBucketReplicationInput struct {
	Role  string            `json:"Role"`
	Rules []ReplicationRule `json:"Rules"`
}

type PutBucketReplicationInput struct {
	Bucket string
	Role   string
	Rules  []ReplicationRule
}

type ReplicationRuleWithProgress struct {
	ReplicationRule
	Progress *Progress `json:"Progress,omitempty"`
}

type ReplicationRule struct {
	ID                          string          `json:"ID"`
	Status                      enum.StatusType `json:"Status"`
	PrefixSet                   []string        `json:"PrefixSet,omitempty"`
	Destination                 Destination     `json:"Destination"`
	HistoricalObjectReplication enum.StatusType `json:"HistoricalObjectReplication"`
}

type Destination struct {
	Bucket                       string                                `json:"Bucket"`
	Location                     string                                `json:"Location"`
	StorageClass                 enum.StorageClassType                 `json:"StorageClass,omitempty"`
	StorageClassInheritDirective enum.StorageClassInheritDirectiveType `json:"StorageClassInheritDirective,omitempty"`
}

type Progress struct {
	HistoricalObject float64 `json:"HistoricalObject"`
	NewObject        string  `json:"NewObject"`
}

type PutBucketReplicationOutput struct {
	RequestInfo
}

type GetBucketReplicationInput struct {
	Bucket string
	RuleID string
}

type GetBucketReplicationOutput struct {
	RequestInfo
	Role  string                        `json:"Role"`
	Rules []ReplicationRuleWithProgress `json:"Rules"`
}
type DeleteBucketReplicationInput struct {
	Bucket string
}

type DeleteBucketReplicationOutput struct {
	RequestInfo
}

type putBucketRealTimeLogInput struct {
	Configuration RealTimeLogConfiguration `json:"RealTimeLogConfiguration"`
}

type PutBucketRealTimeLogInput struct {
	Bucket        string
	Configuration RealTimeLogConfiguration
}

type RealTimeLogConfiguration struct {
	Role          string                 `json:"Role"`
	Configuration AccessLogConfiguration `json:"AccessLogConfiguration"`
}

type AccessLogConfiguration struct {
	UseServiceTopic bool   `json:"UseServiceTopic"`
	TLSProjectID    string `json:"TLSProjectID"`
	TLSTopicID      string `json:"TLSTopicID"`
}

type PutBucketRealTimeLogOutput struct {
	RequestInfo
}

type GetBucketRealTimeLogInput struct {
	Bucket string
}

type GetBucketRealTimeLogOutput struct {
	RequestInfo
	Configuration RealTimeLogConfiguration `json:"RealTimeLogConfiguration"`
}

type DeleteBucketRealTimeLogInput struct {
	Bucket string
}

type DeleteBucketRealTimeLogOutput struct {
	RequestInfo
}
type putBucketCustomDomainInput struct {
	Rule CustomDomainRule `json:"CustomDomainRule,omitempty"`
}

type PutBucketCustomDomainInput struct {
	Bucket string
	Rule   CustomDomainRule
}

type CustomDomainRule struct {
	CertID          string              `json:"CertId"`
	CertStatus      enum.CertStatusType `json:"CertStatus"`
	Domain          string              `json:"Domain"`
	Forbidden       bool                `json:"Forbidden"`
	ForbiddenReason string              `json:"ForbiddenReason"`
	Cname           string              `json:"Cname"`
}

type PutBucketCustomDomainOutput struct {
	RequestInfo
}

type ListBucketCustomDomainInput struct {
	Bucket string
}

type ListBucketCustomDomainOutput struct {
	RequestInfo
	Rules []CustomDomainRule `json:"CustomDomainRules"`
}

type DeleteBucketCustomDomainInput struct {
	Bucket string
	Domain string
}

type DeleteBucketCustomDomainOutput struct {
	RequestInfo
}

type GetBucketEncryptionInput struct {
	Bucket string
}

type PutBucketEncryptionInput struct {
	Bucket string               `json:"-"`
	Rule   BucketEncryptionRule `json:"Rule"`
}

type BucketEncryptionRule struct {
	ApplyServerSideEncryptionByDefault ApplyServerSideEncryptionByDefault `json:"ApplyServerSideEncryptionByDefault"`
}

type ApplyServerSideEncryptionByDefault struct {
	SSEAlgorithm   string `json:"SSEAlgorithm"`
	KMSMasterKeyID string `json:"KMSMasterKeyID,omitempty"`
}

// Response
type PutBucketEncryptionOutput struct {
	RequestInfo
}

type GetBucketEncryptionOutput struct {
	RequestInfo
	Rule BucketEncryptionRule
}

type DeleteBucketEncryptionInput struct {
	Bucket string
}

type DeleteBucketEncryptionOutput struct {
	RequestInfo `json:"-"`
}

type ResumableCopyObjectInput struct {
	CreateMultipartUploadV2Input

	SrcBucket    string
	SrcKey       string
	SrcVersionID string

	CopySourceIfMatch           string
	CopySourceIfModifiedSince   time.Time
	CopySourceIfNoneMatch       string
	CopySourceIfUnmodifiedSince time.Time

	CopySourceSSECAlgorithm string
	CopySourceSSECKey       string
	CopySourceSSECKeyMD5    string

	PartSize         int64
	TaskNum          int
	EnableCheckpoint bool
	CheckpointFile   string
	TrafficLimit     int64

	CopyEventListener CopyEventListener
	CancelHook        CancelHook
}

type CopyPartInfo struct {
	PartNumber           int
	CopySourceRangeStart int64
	CopySourceRangeEnd   int64

	// upload part copy succeed 时有值
	Etag *string
}

type CopyEvent struct {
	Type enum.CopyEventType
	Err  error

	Bucket         string
	Key            string
	UploadID       *string
	SrcBucket      string
	SrcKey         string
	SrcVersionID   string
	CheckpointFile *string
	CopyPartInfo   *copyPartInfo
}

type CopyEventListener interface {
	EventChange(event *CopyEvent)
}

type ResumableCopyObjectOutput struct {
	RequestInfo
	Bucket        string
	Key           string
	UploadID      string
	Etag          string
	Location      string
	VersionID     string
	HashCrc64ecma uint64
	SSECAlgorithm string
	SSECKeyMD5    string
	EncodingType  string
}

type restoreObjectInput struct {
	Days                 int                   `json:"Days"`
	RestoreJobParameters *RestoreJobParameters `json:"RestoreJobParameters,omitempty"`
}
type RestoreObjectInput struct {
	Bucket               string
	Key                  string
	VersionID            string `location:"query" locationName:"versionId"`
	Days                 int
	RestoreJobParameters *RestoreJobParameters
}

type RestoreObjectOutput struct {
	RequestInfo
}
type putBucketTaggingInput struct {
	TagSet TagSet `json:"TagSet"`
}

type PutBucketTaggingInput struct {
	Bucket string
	TagSet TagSet
}

type PutBucketTaggingOutput struct {
	RequestInfo
}

type GetBucketTaggingInput struct {
	Bucket string // required
}

type GetBucketTaggingOutput struct {
	RequestInfo
	TagSet TagSet
}

type DeleteBucketTaggingInput struct {
	Bucket string // required
}

type DeleteBucketTaggingOutput struct {
	RequestInfo
}

type RestoreJobParameters struct {
	Tier enum.TierType `json:"Tier"`
}

type GenericInput struct {
	RequestDate time.Time // 不为空时，代表本次请求 Header 中指定的 X-Tos-Date 头域（转换为 UTC 时间），包含签名时和发送时
	RequestHost string    // 不为空时，代表本次请求 Header 中指定的 Host 头域，仅影响签名和发送请求时的 Host 头域，实际建立仍使用 Endpoint
}

type GetFileStatusInput struct {
	GenericInput        // v2.8.0
	Bucket       string // required
	Key          string // required
}

type GetFileStatusOutput struct {
	RequestInfo
	Key          string
	Size         int64
	LastModified time.Time
	Crc32        string
	Crc64        string
	Etag         string
}

type PutSymlinkInput struct {
	Bucket              string
	Key                 string
	SymlinkTargetKey    string
	SymlinkTargetBucket string                `location:"header" locationName:"X-Tos-Symlink-Bucket"`
	ForbidOverwrite     bool                  `location:"header" locationName:"X-Tos-Forbid-Overwrite"`
	ACL                 enum.ACLType          `location:"header" locationName:"X-Tos-Acl"`
	StorageClass        enum.StorageClassType `location:"header" locationName:"X-Tos-Storage-Class"`
	Tagging             string                `location:"header" locationName:"X-Tos-Tagging"`
	ContentType         string                `location:"header" locationName:"Content-Type"`
	Expires             time.Time             `location:"header" locationName:"Expires"`
	CacheControl        string                `location:"header" locationName:"Cache-Control"`
	ContentDisposition  string                `location:"header" locationName:"Content-Disposition" encodeChinese:"true"`
	ContentLanguage     string                `location:"header" locationName:"Content-Language"`
	Meta                map[string]string     `location:"headers"`
}

type PutSymlinkOutput struct {
	RequestInfo
	VersionID string
}

type GetSymlinkInput struct {
	Bucket    string
	Key       string
	VersionID string
}

type GetSymlinkOutput struct {
	RequestInfo
	VersionID           string
	LastModified        time.Time
	SymlinkTargetKey    string
	SymlinkTargetBucket string
}

type modifyObjectInput struct {
	Bucket  string
	Key     string
	Offset  int64     `location:"query" locationName:"offset" default:"0"`
	Content io.Reader // required

	ContentLength        int64
	DataTransferListener DataTransferListener
	RateLimiter          RateLimiter
	TrafficLimit         int64 `location:"header" locationName:"X-Tos-Traffic-Limit"`
}

type modifyObjectOutput struct {
	RequestInfo
	NextModifyOffset int64
	HashCrc64ecma    uint64
}

type DataTransferListener interface {
	DataTransferStatusChange(status *DataTransferStatus)
}

type RateLimiter interface {
	// Acquire try to get a token.
	// If ok, caller can read want bytes, else wait timeToWait and try again.
	Acquire(want int64) (ok bool, timeToWait time.Duration)
}

type PutQosPolicyInput struct {
	AccountID string
	Policy    string
}

type PutQosPolicyOutput struct {
	RequestInfo
}

type GetQosPolicyInput struct {
	AccountID string
}

type GetQosPolicyOutput struct {
	RequestInfo
	Policy string
}

type DeleteQosPolicyInput struct {
	AccountID string
}

type DeleteQosPolicyOutput struct {
	RequestInfo
}

type GetBucketInfoInput struct {
	Bucket string
}
type GetBucketInfoOutput struct {
	RequestInfo
	Bucket BucketInfo `json:"Bucket"`
}

type BucketInfo struct {
	Name                              string                            `json:"Name"`
	Owner                             Owner                             `json:"Owner"`
	CreationDate                      time.Time                         `json:"CreationDate"`
	StorageClass                      enum.StorageClassType             `json:"StorageClass"`
	ProjectName                       string                            `json:"ProjectName"`
	Type                              enum.BucketType                   `json:"Type"`
	Location                          string                            `json:"Location"`
	AzRedundancy                      enum.AzRedundancyType             `json:"AzRedundancy"`
	ExtranetEndpoint                  string                            `json:"ExtranetEndpoint"`
	IntranetEndpoint                  string                            `json:"IntranetEndpoint"`
	ExtranetS3Endpoint                string                            `json:"ExtranetS3Endpoint"`
	IntranetS3Endpoint                string                            `json:"IntranetS3Endpoint"`
	Versioning                        string                            `json:"Versioning"`
	CrossRegionReplication            enum.StatusType                   `json:"CrossRegionReplication"`
	TransferAcceleration              enum.StatusType                   `json:"TransferAcceleration"`
	AccessMonitor                     enum.StatusType                   `json:"AccessMonitor"`
	ServerSideEncryptionConfiguration ServerSideEncryptionConfiguration `json:"ServerSideEncryptionConfiguration"`
}

type ServerSideEncryptionConfiguration struct {
	Rule BucketEncryptionRule `json:"Rule"`
}
