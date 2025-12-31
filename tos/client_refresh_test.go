package tos

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randomID2() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func TestClientV2_RefreshEndpointRegion(t *testing.T) {
	endpoint := "tos-cn-beijing.volces.com"
	region := "cn-beijing"
	cred := NewStaticCredentials(randomID2(), randomID2())
	cli, err := NewClientV2(endpoint, WithRegion(region), WithCredentials(cred))
	assert.Nil(t, err)

	// Test 1: Update Endpoint
	newEndpoint := "tos-cn-shanghai.volces.com"
	success := cli.RefreshEndpointRegion(newEndpoint, "")
	assert.True(t, success)
	assert.Equal(t, newEndpoint, cli.config.Endpoint)
	assert.Equal(t, "https", cli.scheme) // Default scheme
	assert.Equal(t, newEndpoint, cli.host)

	// Test 2: Update Region
	newRegion := "cn-shanghai"
	success = cli.RefreshEndpointRegion("", newRegion)
	assert.True(t, success)
	assert.Equal(t, newRegion, cli.config.Region)
	// Check if signer region is updated
	signer, ok := cli.signer.(*SignV4)
	assert.True(t, ok)
	assert.Equal(t, newRegion, signer.region)

	// Test 3: Update Both
	newEndpoint2 := "tos-cn-guangzhou.volces.com"
	newRegion2 := "cn-guangzhou"
	success = cli.RefreshEndpointRegion(newEndpoint2, newRegion2)
	assert.True(t, success)
	assert.Equal(t, newEndpoint2, cli.config.Endpoint)
	assert.Equal(t, newRegion2, cli.config.Region)
	signer, ok = cli.signer.(*SignV4)
	assert.True(t, ok)
	assert.Equal(t, newRegion2, signer.region)

	// Test 4: Invalid Endpoint
	invalidEndpoint := "tos-s3-cn-beijing.volces.com"
	success = cli.RefreshEndpointRegion(invalidEndpoint, "")
	assert.False(t, success)
	// Should remain unchanged
	assert.Equal(t, newEndpoint2, cli.config.Endpoint)

	// Test 5: Empty Input
	success = cli.RefreshEndpointRegion("", "")
	assert.False(t, success)
}
