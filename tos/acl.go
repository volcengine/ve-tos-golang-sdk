package tos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Grantee struct {
	ID          string `json:"ID,omitempty"`
	DisplayName string `json:"DisplayName,omitempty"`
	Type        string `json:"Type,omitempty"`
	URI         string `json:"Canned,omitempty"`
}

type Grant struct {
	Grantee    Grantee `json:"Grantee,omitempty"`
	Permission string  `json:"Permission,omitempty"`
}

type ObjectAclGrant struct {
	ACL              string `json:"ACL,omitempty"`
	GrantFullControl string `json:"GrantFullControl,omitempty"`
	GrantRead        string `json:"GrantRead,omitempty"`
	GrantReadAcp     string `json:"GrantReadAcp,omitempty"`
	GrantWrite       string `json:"GrantWrite,omitempty"`
	GrantWriteAcp    string `json:"GrantWriteAcp,omitempty"`
}

type ObjectAclRules struct {
	Owner  Owner   `json:"Owner,omitempty"`
	Grants []Grant `json:"Grants,omitempty"`
}

// PutObjectAclInput AclGrant, AclRules can not set both.
type PutObjectAclInput struct {
	Key       string          `json:"Key,omitempty"`       // the object, required
	VersionID string          `json:"VersionId,omitempty"` // the version id of the object, optional
	AclGrant  *ObjectAclGrant `json:"AclGrant,omitempty"`  // set acl by header
	AclRules  *ObjectAclRules `json:"AclRules,omitempty"`  // set acl by rules
}

type PutObjectAclOutput struct {
	RequestInfo `json:"-"`
}

// PutObjectAcl PutObjectAclInput AclGrant, AclRules can not set both.
func (bkt *Bucket) PutObjectAcl(ctx context.Context, input *PutObjectAclInput, options ...Option) (*PutObjectAclOutput, error) {
	if err := isValidKey(input.Key); err != nil {
		return nil, err
	}

	var content io.Reader
	if input.AclRules != nil {
		data, err := json.Marshal(input.AclRules)
		if err != nil {
			return nil, fmt.Errorf("tos: marshal BucketAcl Ruels err: %s", err.Error())
		}
		content = bytes.NewReader(data)
	}

	builder := bkt.client.newBuilder(bkt.name, input.Key, options...).
		WithQuery("acl", "")
	if grant := input.AclGrant; grant != nil {
		builder.WithHeader(HeaderACL, grant.ACL).
			WithHeader(HeaderGrantFullControl, grant.GrantFullControl).
			WithHeader(HeaderGrantRead, grant.GrantRead).
			WithHeader(HeaderGrantReadAcp, grant.GrantReadAcp).
			WithHeader(HeaderGrantWrite, grant.GrantWrite).
			WithHeader(HeaderGrantWriteAcp, grant.GrantWriteAcp)
	}

	res, err := builder.Request(ctx, http.MethodPut, content, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &PutObjectAclOutput{RequestInfo: res.RequestInfo()}, nil
}

type GetObjectAclOutput struct {
	RequestInfo `json:"-"`
	VersionID   string  `json:"VersionId,omitempty"`
	Owner       Owner   `json:"Owner,omitempty"`
	Grants      []Grant `json:"Grants,omitempty"`
}

// GetObjectAcl get object acl
//   objectKey: the name of object
//   Options: WithVersionID the version of the object
func (bkt *Bucket) GetObjectAcl(ctx context.Context, objectKey string, options ...Option) (*GetObjectAclOutput, error) {
	if err := isValidKey(objectKey); err != nil {
		return nil, err
	}

	res, err := bkt.client.newBuilder(bkt.name, objectKey, options...).
		WithQuery("acl", "").
		Request(ctx, http.MethodGet, nil, bkt.client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	out := GetObjectAclOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(out.RequestID, res.Body, &out); err != nil {
		return nil, err
	}

	out.VersionID = res.Header.Get(HeaderVersionID)
	return &out, nil
}
