package tos

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func TestClientV2_RefreshCredentials(t *testing.T) {
	endpoint := "tos-cn-beijing.volces.com"
	region := "cn-beijing"
	accessKey := randomID()
	secretKey := randomID()
	cred := NewStaticCredentials(accessKey, secretKey)
	cli, err := NewClientV2(endpoint, WithRegion(region), WithCredentials(cred))
	assert.Nil(t, err)

	// Test 1: Update AK/SK
	newAccessKey := randomID()
	newSecretKey := randomID()
	success := cli.RefreshCredentials(newAccessKey, newSecretKey, "")
	assert.True(t, success)

	// Check if credentials updated
	c := cli.credentials.Credential()
	assert.Equal(t, newAccessKey, c.AccessKeyID)
	assert.Equal(t, newSecretKey, c.AccessKeySecret)
	assert.Empty(t, c.SecurityToken)

	// Check if signer updated
	signer, ok := cli.signer.(*SignV4)
	assert.True(t, ok)
	assert.Equal(t, newAccessKey, signer.credentials.Credential().AccessKeyID)

	// Test 2: Update with SecurityToken
	newAccessKey2 := randomID()
	newSecretKey2 := randomID()
	token := randomID()
	success = cli.RefreshCredentials(newAccessKey2, newSecretKey2, token)
	assert.True(t, success)

	c = cli.credentials.Credential()
	assert.Equal(t, newAccessKey2, c.AccessKeyID)
	assert.Equal(t, newSecretKey2, c.AccessKeySecret)
	assert.Equal(t, token, c.SecurityToken)

	signer, ok = cli.signer.(*SignV4)
	assert.True(t, ok)
	assert.Equal(t, newAccessKey2, signer.credentials.Credential().AccessKeyID)
	assert.Equal(t, token, signer.credentials.Credential().SecurityToken)

	// Test 3: Empty AK or SK
	success = cli.RefreshCredentials("", randomID(), "")
	assert.False(t, success)
	success = cli.RefreshCredentials(randomID(), "", "")
	assert.False(t, success)

	// Values should remain unchanged
	c = cli.credentials.Credential()
	assert.Equal(t, newAccessKey2, c.AccessKeyID)
}
