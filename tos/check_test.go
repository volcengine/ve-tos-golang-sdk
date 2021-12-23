package tos

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidBucketName(t *testing.T) {
	err := IsValidBucketName("-x-")
	require.NotNil(t, err)

	err = IsValidBucketName("x")
	require.NotNil(t, err)

	err = IsValidBucketName("xx😊xx")
	require.NotNil(t, err)

	err = IsValidBucketName("xx😊xx")
	require.NotNil(t, err)

	name := strings.Repeat("a", 100)
	err = IsValidBucketName(name)
	require.NotNil(t, err)

	err = IsValidBucketName("abc123")
	require.Nil(t, err)
}
