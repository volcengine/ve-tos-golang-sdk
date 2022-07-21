package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/tos"
	"github.com/volcengine/ve-tos-golang-sdk/tos/enum"
)

func TestObjectACLV2(t *testing.T) {
	var (
		env     = newTestEnv(t)
		bucket  = generateBucketName("put-object-acl")
		client  = env.prepareClient(bucket)
		key     = "key123"
		ownerID = "test-owner-id"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	putRandomObject(t, client, bucket, key, 4*1024)
	acl, err := client.PutObjectACL(context.Background(), &tos.PutObjectACLInput{
		Bucket: bucket,
		Key:    key,
		Grants: []tos.Grant{{
			Grantee: tos.Grantee{
				ID:   ownerID,
				Type: "CanonicalUser",
			},
			Permission: enum.PermissionRead,
		}},
	})
	require.Nil(t, err)
	require.Equal(t, 200, acl.StatusCode)
	getAcl, err := client.GetObjectACL(context.Background(), &tos.GetObjectACLInput{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, 200, getAcl.StatusCode)
	require.Equal(t, ownerID, getAcl.Grants[0].Grantee.ID)
	require.Equal(t, enum.PermissionRead, getAcl.Grants[0].Permission)
}

func TestPutWithACLV2(t *testing.T) {
	var (
		env     = newTestEnv(t)
		bucket  = generateBucketName("put-object-acl")
		client  = env.prepareClient(bucket)
		key     = "key123"
		content = "value123"
		ownerID = "test-owner-id"
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{
			Bucket:    bucket,
			Key:       key,
			GrantRead: "id=" + ownerID},
		Content: strings.NewReader(content),
	})
	checkSuccess(t, put, err, 200)
	getAcl, err := client.GetObjectACL(context.Background(), &tos.GetObjectACLInput{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, 200, getAcl.StatusCode)
	require.Equal(t, ownerID, getAcl.Grants[0].Grantee.ID)
}
