package tests

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestNewClient(t *testing.T) {
	client, err := tos.NewClientV2("tos-s3-cn-shanghai.volces.com")
	require.Equal(t, err, tos.InvalidS3Endpoint)
	require.Nil(t, client)
}

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
		Grants: []tos.GrantV2{{
			GranteeV2: tos.GranteeV2{
				ID:   ownerID,
				Type: "CanonicalUser",
			},
			Permission: enum.PermissionRead,
		}},
		BucketOwnerEntrusted: true,
	})
	require.Nil(t, err)
	require.Equal(t, 200, acl.StatusCode)
	getAcl, err := client.GetObjectACL(context.Background(), &tos.GetObjectACLInput{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, 200, getAcl.StatusCode)
	require.Equal(t, ownerID, getAcl.Grants[0].GranteeV2.ID)
	require.Equal(t, enum.PermissionRead, getAcl.Grants[0].Permission)
	require.Equal(t, getAcl.BucketOwnerEntrusted, true)
	ctx := context.Background()
	acl, err = client.PutObjectACL(ctx, &tos.PutObjectACLInput{
		Bucket:     bucket,
		Key:        key,
		GrantWrite: "id=123",
	})
	require.Nil(t, err)

	getAcl, err = client.GetObjectACL(context.Background(), &tos.GetObjectACLInput{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, len(getAcl.Grants), 1)
	require.Equal(t, getAcl.Grants[0].GranteeV2.ID, "123")
	require.Equal(t, getAcl.Grants[0].Permission, enum.PermissionWrite)

	acl, err = client.PutObjectACL(context.Background(), &tos.PutObjectACLInput{
		Bucket: bucket,
		Key:    key,
		Grants: []tos.GrantV2{{
			GranteeV2: tos.GranteeV2{
				ID:   ownerID,
				Type: "CanonicalUser",
			},
			Permission: enum.PermissionRead,
		}},
		BucketOwnerEntrusted: false,
	})
	require.Nil(t, err)
	require.Equal(t, 200, acl.StatusCode)
	getAcl, err = client.GetObjectACL(context.Background(), &tos.GetObjectACLInput{
		Bucket: bucket,
		Key:    key,
	})
	require.Nil(t, err)
	require.Equal(t, 200, getAcl.StatusCode)
	require.Equal(t, ownerID, getAcl.Grants[0].GranteeV2.ID)
	require.Equal(t, enum.PermissionRead, getAcl.Grants[0].Permission)
	require.Equal(t, getAcl.BucketOwnerEntrusted, false)
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
	require.Equal(t, ownerID, getAcl.Grants[0].GranteeV2.ID)
}

func TestPutObjectV2WithACL(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-bucket-acl")
		client = env.prepareClient(bucket)
		key    = "test"
		value  = randomString(1024)
		ctx    = context.Background()
	)
	_, err := client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ACL: enum.ACLBucketOwnerFullControl},
		Content:             bytes.NewBufferString(value),
	})
	require.Nil(t, err)
	out, err := client.GetObjectACL(ctx, &tos.GetObjectACLInput{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, len(out.Grants), 1)
	require.Equal(t, out.BucketOwnerEntrusted, false)

	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key, ACL: enum.ACLBucketOwnerEntrusted},
		Content:             bytes.NewBufferString(value),
	})
	require.Nil(t, err)
	out, err = client.GetObjectACL(ctx, &tos.GetObjectACLInput{Bucket: bucket, Key: key})
	require.Nil(t, err)
	require.Equal(t, out.BucketOwnerEntrusted, true)

	newKey := "default-" + randomString(8)
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: newKey},
		Content:             nil,
	})
	require.Nil(t, err)
	_, err = client.PutObjectACL(ctx, &tos.PutObjectACLInput{
		Bucket:    bucket,
		Key:       newKey,
		IsDefault: true,
	})
	require.Nil(t, err)

	aclOut, err := client.GetObjectACL(ctx, &tos.GetObjectACLInput{Bucket: bucket, Key: newKey})
	require.Nil(t, err)
	require.Equal(t, aclOut.IsDefault, true)

	newKey = "default-" + randomString(8)
	_, err = client.PutObjectV2(ctx, &tos.PutObjectV2Input{
		PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: newKey, ACL: enum.ACLDefault},
		Content:             nil,
	})
	require.Nil(t, err)

	aclOut, err = client.GetObjectACL(ctx, &tos.GetObjectACLInput{Bucket: bucket, Key: newKey})
	require.Nil(t, err)
	require.Equal(t, aclOut.IsDefault, true)
}

func TestBucketACLGrantsBody(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-bucket-acl")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	displayName := "displayName"
	grantAccountId := "123"
	res, err := client.GetBucketACL(ctx, &tos.GetBucketACLInput{Bucket: bucket})
	require.Nil(t, err)
	accountId := res.Owner.ID
	grant := []tos.GrantV2{{
		GranteeV2:  tos.GranteeV2{ID: accountId, Type: enum.GranteeUser},
		Permission: enum.PermissionFullControl,
	}, {
		GranteeV2:  tos.GranteeV2{ID: grantAccountId, DisplayName: displayName, Type: enum.GranteeUser},
		Permission: enum.PermissionRead,
	}, {
		GranteeV2:  tos.GranteeV2{ID: grantAccountId, Type: enum.GranteeUser},
		Permission: enum.PermissionWrite,
	}, {
		GranteeV2:  tos.GranteeV2{Type: enum.GranteeGroup, Canned: enum.CannedAllUsers},
		Permission: enum.PermissionReadAcp,
	}, {
		GranteeV2:  tos.GranteeV2{Type: enum.GranteeGroup, Canned: enum.CannedAuthenticatedUsers},
		Permission: enum.PermissionWriteAcp,
	}}

	_, err = client.PutBucketACL(ctx, &tos.PutBucketACLInput{
		Bucket: bucket,
		Owner: tos.Owner{
			ID: accountId,
		},
		Grants: grant})
	require.Nil(t, err)

	res, err = client.GetBucketACL(ctx, &tos.GetBucketACLInput{Bucket: bucket})
	assert.Nil(t, err)
	assert.Equal(t, len(res.Grants), len(grant))
	for _, g := range res.Grants {
		assertAlready := false
		for _, rowG := range grant {
			if g.Permission == rowG.Permission {
				assertAlready = true
				assert.Equal(t, g, rowG)
				break
			}
		}
		assert.True(t, assertAlready)
	}
}

func TestBucketACLGrantHeader(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-bucket-acl")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()

	_, err := client.PutBucketACL(ctx, &tos.PutBucketACLInput{
		Bucket:           bucket,
		GrantFullControl: "id=123",
		GrantRead:        "id=123",
		GrantReadAcp:     "id=123",
		GrantWrite:       "id=123",
		GrantWriteAcp:    "id=123",
	})
	require.Nil(t, err)
	bucketACL, err := client.GetBucketACL(ctx, &tos.GetBucketACLInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.Equal(t, len(bucketACL.Grants), 5)
}

func TestBucketACL(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-bucket-acl")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	inputACLTypeList := []enum.ACLType{
		enum.ACLPrivate,
		enum.ACLPublicRead,
		enum.ACLPublicReadWrite,
		enum.ACLAuthRead,
		enum.ACLBucketOwnerRead,
		enum.ACLBucketOwnerFullControl}
	wantRes := []struct {
		Type   enum.GranteeType
		Canned enum.CannedType
	}{{
		Type: enum.GranteeUser, Canned: ""},
		{enum.GranteeGroup, enum.CannedAllUsers},
		{enum.GranteeGroup, enum.CannedAllUsers},
		{enum.GranteeGroup, enum.CannedAuthenticatedUsers},
		{enum.GranteeUser, ""},
		{enum.GranteeUser, ""},
	}
	for i, inputAcl := range inputACLTypeList {
		_, err := client.PutBucketACL(ctx, &tos.PutBucketACLInput{
			Bucket:  bucket,
			ACLType: inputAcl,
		})
		require.Nil(t, err)
		bucketACL, err := client.GetBucketACL(ctx, &tos.GetBucketACLInput{
			Bucket: bucket,
		})
		require.Nil(t, err)
		if inputAcl == enum.ACLPublicReadWrite {
			require.Equal(t, len(bucketACL.Grants), 2)

		} else {
			require.Equal(t, len(bucketACL.Grants), 1)
		}

		require.Equal(t, bucketACL.Grants[0].GranteeV2.Canned, wantRes[i].Canned)
		require.Equal(t, bucketACL.Grants[0].GranteeV2.Type, wantRes[i].Type)
	}

	_, err := client.PutBucketACL(ctx, &tos.PutBucketACLInput{Bucket: bucket, BucketAclDelivered: true})
	require.Nil(t, err)

	out, err := client.GetBucketACL(ctx, &tos.GetBucketACLInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, out.BucketAclDelivered, true)

	cli, err := tos.NewClient(env.endpoint, tos.WithCredentials(tos.NewStaticCredentials(env.accessKey, env.secretKey)), tos.WithRegion(env.region))
	require.Nil(t, err)
	bucketAcl, err := cli.GetBucketACL(ctx, &tos.GetBucketACLInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, bucketAcl.BucketAclDelivered, true)

	_, err = cli.PutBucketACL(ctx, &tos.PutBucketACLInput{
		Bucket:  bucket,
		ACLType: enum.ACLPublicRead,
	})
	require.Nil(t, err)
	bucketAcl, err = cli.GetBucketACL(ctx, &tos.GetBucketACLInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, bucketAcl.BucketAclDelivered, false)

}

func TestBucketAcl(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("put-bucket-acl")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()

	_, err := client.PutBucketACL(ctx, &tos.PutBucketACLInput{Bucket: bucket, BucketAclDelivered: true})
	require.Nil(t, err)

	out, err := client.GetBucketACL(ctx, &tos.GetBucketACLInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, out.BucketAclDelivered, true)
}

func TestObjectACLV1(t *testing.T) {
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
	clientV1, err := tos.NewClient(env.endpoint, tos.WithCredentials(tos.NewStaticCredentials(env.accessKey, env.secretKey)), tos.WithRegion(env.region))
	bkt, err := clientV1.Bucket(bucket)
	require.Nil(t, err)
	acl, err := bkt.PutObjectAcl(context.Background(), &tos.PutObjectAclInput{
		Key: key,
		AclRules: &tos.ObjectAclRules{
			Owner: tos.Owner{},
			Grants: []tos.Grant{{
				Grantee: tos.Grantee{
					ID:   ownerID,
					Type: "CanonicalUser",
				},
				Permission: enum.PermissionRead,
			}},
			BucketOwnerEntrusted: true,
		},
	})
	require.Nil(t, err)
	require.Equal(t, 200, acl.StatusCode)
	getAcl, err := bkt.GetObjectAcl(context.Background(), key)
	require.Nil(t, err)
	require.Equal(t, 200, getAcl.StatusCode)
	require.Equal(t, ownerID, getAcl.Grants[0].Grantee.ID)
	require.Equal(t, enum.PermissionRead, getAcl.Grants[0].Permission)
	require.Equal(t, getAcl.BucketOwnerEntrusted, true)
	ctx := context.Background()
	acl, err = bkt.PutObjectAcl(ctx, &tos.PutObjectAclInput{
		Key:      key,
		AclGrant: &tos.ObjectAclGrant{GrantRead: "id=123"},
	})
	require.Nil(t, err)

	getAcl, err = bkt.GetObjectAcl(context.Background(), key)
	require.Nil(t, err)
	require.Equal(t, len(getAcl.Grants), 1)
	require.Equal(t, getAcl.Grants[0].Grantee.ID, "123")
	require.Equal(t, getAcl.Grants[0].Permission, enum.PermissionRead)

	acl, err = bkt.PutObjectAcl(context.Background(), &tos.PutObjectAclInput{
		Key: key,
		AclRules: &tos.ObjectAclRules{
			Owner: tos.Owner{},
			Grants: []tos.Grant{{
				Grantee: tos.Grantee{
					ID:   ownerID,
					Type: "CanonicalUser",
				}, Permission: enum.PermissionRead}},
			IsDefault:            false,
			BucketOwnerEntrusted: true,
		},
	})
	require.Nil(t, err)
	require.Equal(t, 200, acl.StatusCode)
	getAcl, err = bkt.GetObjectAcl(context.Background(),
		key,
	)
	require.Nil(t, err)
	require.Equal(t, 200, getAcl.StatusCode)
	require.Equal(t, ownerID, getAcl.Grants[0].Grantee.ID)
	require.Equal(t, enum.PermissionRead, getAcl.Grants[0].Permission)
	require.Equal(t, getAcl.BucketOwnerEntrusted, true)
	require.Equal(t, getAcl.IsDefault, false)

	newkey := randomString(8)
	putRandomObject(t, client, bucket, newkey, 4*1024)

	acl, err = bkt.PutObjectAcl(context.Background(), &tos.PutObjectAclInput{
		Key: newkey,
		AclRules: &tos.ObjectAclRules{
			Owner: tos.Owner{
				ID:          ownerID,
				DisplayName: "123",
			},
			IsDefault:            true,
			BucketOwnerEntrusted: false,
		},
	})
	require.Nil(t, err)
	require.Equal(t, 200, acl.StatusCode)

	getAcl, err = bkt.GetObjectAcl(context.Background(),
		newkey,
	)
	require.Nil(t, err)
	require.Equal(t, 200, getAcl.StatusCode)
	require.Equal(t, getAcl.BucketOwnerEntrusted, false)
	require.Equal(t, getAcl.IsDefault, true)
}
