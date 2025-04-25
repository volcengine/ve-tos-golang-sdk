package tos

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

type RestoreInfo struct {
	RestoreStatus RestoreStatus
	RestoreParam  *RestoreParam // GetObject 接口当前该值为空
}

type RestoreStatus struct {
	OngoingRequest bool
	ExpiryDate     time.Time
}

type RestoreParam struct {
	RequestDate time.Time
	ExpiryDays  int
	Tier        enum.TierType
}

// ObjectMeta object metadata
type ObjectMeta struct {
	ContentLength        int64             `json:"ContentLength,omitempty"`
	ContentType          string            `json:"ContentType,omitempty"`
	ContentMD5           string            `json:"ContentMD5,omitempty"`
	ContentLanguage      string            `json:"ContentLanguage,omitempty"`
	ContentEncoding      string            `json:"ContentEncoding,omitempty"`
	ContentDisposition   string            `json:"ContentDisposition,omitempty"`
	LastModified         string            `json:"LastModified,omitempty"`
	CacheControl         string            `json:"CacheControl,omitempty"`
	Expires              string            `json:"Expires,omitempty"`
	ETag                 string            `json:"ETag,omitempty"`
	VersionID            string            `json:"VersionId,omitempty"`
	DeleteMarker         bool              `json:"DeleteMarker,omitempty"`
	ObjectType           string            `json:"ObjectType,omitempty"` // "" or "Appendable"
	StorageClass         string            `json:"StorageClass,omitempty"`
	RestoreInfo          *RestoreInfo      `json:"Restore,omitempty"`
	Metadata             map[string]string `json:"Metadata,omitempty"`
	Tag                  string            `json:"Tag,omitempty"`
	SSECustomerAlgorithm string            `json:"SSECustomerAlgorithm,omitempty"`
	SSECustomerKeyMD5    string            `json:"SSECustomerKeyMD5,omitempty"`
	CSType               string            `json:"CSType,omitempty"`
	IsDirectory          bool              `json:"IsDirectory"`
}

type ObjectMetaV2 struct {
	ETag                      string
	LastModified              time.Time
	DeleteMarker              bool
	SSECAlgorithm             string
	SSECKeyMD5                string
	VersionID                 string
	WebsiteRedirectLocation   string
	ObjectType                string
	HashCrc64ecma             uint64
	StorageClass              enum.StorageClassType
	Meta                      Metadata
	ContentLength             int64
	ContentType               string
	CacheControl              string
	ContentDisposition        string
	ContentEncoding           string
	ContentLanguage           string
	Expires                   time.Time
	ServerSideEncryption      string
	ServerSideEncryptionKeyID string
	ReplicationStatus         enum.ReplicationStatusType
	RestoreInfo               *RestoreInfo `json:"Restore,omitempty"`
	IsDirectory               bool
	Expiration                string
	TaggingCount              int64
}

func parseRestoreInfo(res *Response) *RestoreInfo {
	restore := res.Header.Get(HeaderRestore)
	if restore == "" {
		return nil
	}
	resp := &RestoreInfo{}
	param := parseRestoreParams(restore)
	if result, ok := param["ongoing-request"]; ok && result == "true" {
		resp.RestoreStatus.OngoingRequest = true
	}
	if param["expiry-date"] != "" {
		expiryDate, _ := time.ParseInLocation(http.TimeFormat, param["expiry-date"], time.UTC)
		resp.RestoreStatus.ExpiryDate = expiryDate
	}

	if res.Header.Get(HeaderRestoreRequestDate) != "" || res.Header.Get(HeaderRestoreTier) != "" || res.Header.Get(HeaderRestoreExpiryDays) != "" {
		resp.RestoreParam = &RestoreParam{}
	}

	if res.Header.Get(HeaderRestoreRequestDate) != "" {
		requestDate, _ := time.ParseInLocation(http.TimeFormat, res.Header.Get(HeaderRestoreRequestDate), time.UTC)
		resp.RestoreParam.RequestDate = requestDate

	}

	if res.Header.Get(HeaderRestoreExpiryDays) != "" {
		expiryDays := res.Header.Get(HeaderRestoreExpiryDays)
		expiryDay, err := strconv.ParseInt(expiryDays, 10, 64)
		if err == nil {
			resp.RestoreParam.ExpiryDays = int(expiryDay)

		}
	}

	if res.Header.Get(HeaderRestoreTier) != "" {
		resp.RestoreParam.Tier = enum.TierType(res.Header.Get(HeaderRestoreTier))
	}
	return resp
}

func (om *ObjectMeta) fromResponse(res *Response, disableEncodingMeta bool) {
	om.ETag = res.Header.Get(HeaderETag)
	om.LastModified = res.Header.Get(HeaderLastModified)
	om.DeleteMarker, _ = strconv.ParseBool(res.Header.Get(HeaderDeleteMarker))
	om.SSECustomerAlgorithm = res.Header.Get(HeaderSSECustomerAlgorithm)
	om.SSECustomerKeyMD5 = res.Header.Get(HeaderSSECustomerKeyMD5)
	om.VersionID = res.Header.Get(HeaderVersionID)

	om.ObjectType = res.Header.Get(HeaderObjectType)
	om.StorageClass = res.Header.Get(HeaderStorageClass)
	om.Metadata = userMetadata(res.Header, disableEncodingMeta)

	om.ContentLength = res.ContentLength
	om.ContentType = res.Header.Get(HeaderContentType)
	om.CacheControl = res.Header.Get(HeaderCacheControl)
	contentDisposition, err := url.QueryUnescape(res.Header.Get(HeaderContentDisposition))
	if disableEncodingMeta || err != nil {
		contentDisposition = res.Header.Get(HeaderContentDisposition)
	}
	om.ContentDisposition = contentDisposition
	om.ContentEncoding = res.Header.Get(HeaderContentEncoding)
	om.ContentLanguage = res.Header.Get(HeaderContentLanguage)
	om.Expires = res.Header.Get(HeaderExpires)

	om.ContentMD5 = res.Header.Get(HeaderContentMD5)
	om.Tag = res.Header.Get(HeaderTag)
	om.CSType = res.Header.Get(HeaderCSType)
	om.RestoreInfo = parseRestoreInfo(res)
	om.IsDirectory = res.Header.Get(HeaderDirectory) == "true"

}
func parseRestoreParams(params string) map[string]string {
	result := make(map[string]string)
	parts := strings.SplitAfterN(params, ",", 2) // 按逗号分割参数

	for _, part := range parts {
		keyValue := strings.Split(part, "=") // 按等号分割键和值
		if len(keyValue) == 2 {
			key := strings.TrimSpace(keyValue[0])
			value := strings.Trim(keyValue[1], `"`) // 去除可能存在的引号
			result[key] = value
		}
	}

	return result
}

func (om *ObjectMetaV2) fromResponseV2(res *Response, disableEncodingMeta bool) {
	lastModified, _ := time.ParseInLocation(http.TimeFormat, res.Header.Get(HeaderLastModified), time.UTC)
	deleteMarker, _ := strconv.ParseBool(res.Header.Get(HeaderDeleteMarker))
	// If s is empty or contains invalid digits, err.Err = ErrSyntax and the returned value is 0;
	crc64, _ := strconv.ParseUint(res.Header.Get(HeaderHashCrc64ecma), 10, 64)
	length, _ := strconv.ParseInt(res.Header.Get(HeaderContentLength), 10, 64)
	expires, _ := time.ParseInLocation(http.TimeFormat, res.Header.Get(HeaderExpires), time.UTC)
	om.ETag = res.Header.Get(HeaderETag)
	om.LastModified = lastModified
	om.DeleteMarker = deleteMarker
	om.SSECAlgorithm = res.Header.Get(HeaderSSECustomerAlgorithm)
	om.SSECKeyMD5 = res.Header.Get(HeaderContentMD5)
	om.VersionID = res.Header.Get(HeaderVersionID)
	om.WebsiteRedirectLocation = res.Header.Get(HeaderWebsiteRedirectLocation)
	om.ObjectType = res.Header.Get(HeaderObjectType)
	om.HashCrc64ecma = crc64
	om.StorageClass = enum.StorageClassType(res.Header.Get(HeaderStorageClass))
	om.Meta = &CustomMeta{m: userMetadata(res.Header, disableEncodingMeta)}
	om.ContentLength = length
	om.ContentType = res.Header.Get(HeaderContentType)
	om.CacheControl = res.Header.Get(HeaderCacheControl)
	contentDisposition, err := url.QueryUnescape(res.Header.Get(HeaderContentDisposition))
	if disableEncodingMeta || err != nil {
		contentDisposition = res.Header.Get(HeaderContentDisposition)
	}
	om.ContentDisposition = contentDisposition
	om.ContentEncoding = strings.TrimPrefix(res.Header.Get(HeaderContentEncoding), "tos-raw-trailer,")
	om.ContentLanguage = res.Header.Get(HeaderContentLanguage)
	om.Expires = expires
	om.ServerSideEncryption = res.Header.Get(HeaderServerSideEncryption)
	om.ServerSideEncryptionKeyID = res.Header.Get(HeaderServerSideEncryptionKmsKeyID)
	om.ReplicationStatus = enum.ReplicationStatusType(res.Header.Get(HeaderReplicationStatus))
	om.RestoreInfo = parseRestoreInfo(res)
	om.IsDirectory = res.Header.Get(HeaderDirectory) == "true"
	om.Expiration = res.Header.Get(HeaderTosExpiration)

	if taggingCountHeader := res.Header.Get(HeaderTaggingCount); taggingCountHeader != "" {
		taggingCount, _ := strconv.ParseInt(taggingCountHeader, 10, 64)
		om.TaggingCount = taggingCount
	}
	if res.Header.Get(HeaderRawContentLength) != "" {
		length, err := strconv.ParseInt(res.Header.Get(HeaderRawContentLength), 10, 64)
		if err == nil {
			om.ContentLength = length
		}
	}
}

func userMetadata(header http.Header, disableEncodingMeta bool) map[string]string {

	meta := make(map[string]string)
	for key := range header {
		if strings.HasPrefix(key, HeaderMetaPrefix) {
			if disableEncodingMeta {
				meta[strings.ToLower(key[len(HeaderMetaPrefix):])] = header.Get(key)
				continue
			}
			kk, err := url.QueryUnescape(key[len(HeaderMetaPrefix):])
			if err != nil {
				kk = key[len(HeaderMetaPrefix):]
			}
			meta[strings.ToLower(kk)], err = url.QueryUnescape(header.Get(key))
			if err != nil {
				meta[strings.ToLower(kk)] = header.Get(key)
			}
		}
	}
	return meta
}

func parseUserMetaData(userMeta []userMeta) Metadata {
	if len(userMeta) == 0 {
		return nil
	}
	metas := make(map[string]string, len(userMeta))
	for _, meta := range userMeta {
		kk, err := url.QueryUnescape(meta.Key)
		if err != nil {
			kk = meta.Key
		}
		metas[strings.ToLower(kk)], err = url.QueryUnescape(meta.Value)
		if err != nil {
			metas[strings.ToLower(kk)] = meta.Value
		}
	}
	return &CustomMeta{metas}
}
