package tos

import (
	"net/http"
	"strconv"
	"strings"
)

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
	Restore              string            `json:"Restore,omitempty"`
	Metadata             map[string]string `json:"Metadata,omitempty"`
	Tag                  string            `json:"Tag,omitempty"`
	SSECustomerAlgorithm string            `json:"SSECustomerAlgorithm,omitempty"`
	SSECustomerKeyMD5    string            `json:"SSECustomerKeyMD5,omitempty"`
	CSType               string            `json:"CSType,omitempty"`
}

func (om *ObjectMeta) fromResponse(res *Response) {
	om.ContentLength = res.ContentLength
	om.ContentType = res.Header.Get(HeaderContentType)
	om.ContentMD5 = res.Header.Get(HeaderContentMD5)
	om.ContentLanguage = res.Header.Get(HeaderContentLanguage)
	om.ContentEncoding = res.Header.Get(HeaderContentEncoding)
	om.ContentDisposition = res.Header.Get(HeaderContentDisposition)
	om.LastModified = res.Header.Get(HeaderLastModified)
	om.CacheControl = res.Header.Get(HeaderCacheControl)
	om.Expires = res.Header.Get(HeaderExpires)
	om.ETag = res.Header.Get(HeaderETag)
	om.VersionID = res.Header.Get(HeaderVersionID)
	om.DeleteMarker, _ = strconv.ParseBool(res.Header.Get(HeaderDeleteMarker))
	om.ObjectType = res.Header.Get(HeaderObjectType)
	om.StorageClass = res.Header.Get(HeaderStorageClass)
	om.Restore = res.Header.Get(HeaderRestore)
	om.Metadata = userMetadata(res.Header)
	om.Tag = res.Header.Get(HeaderTag)
	om.SSECustomerAlgorithm = res.Header.Get(HeaderSSECustomerAlgorithm)
	om.SSECustomerKeyMD5 = res.Header.Get(HeaderSSECustomerKeyMD5)
	om.CSType = res.Header.Get(HeaderCSType)
}

func userMetadata(header http.Header) map[string]string {
	meta := make(map[string]string)
	for key := range header {
		if strings.HasPrefix(key, HeaderMetaPrefix) {
			kk := key[len(HeaderMetaPrefix):]
			meta[kk] = header.Get(key)
		}
	}
	return meta
}
