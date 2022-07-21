package tos

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

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

func utf8ToGbk(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func TestIsValidObjectKey(t *testing.T) {
	err := isValidKey("key")
	require.Nil(t, err)

	// utf-8 encode
	err = isValidKey("中文测试")
	require.Nil(t, err)

	err = isValidKey("/key")
	require.NotNil(t, err)

	err = isValidKey("\\key")
	require.NotNil(t, err)

	longKey := make([]byte, 696)
	for i := 0; i < len(longKey); i++ {
		longKey[i] = 32
	}
	err = isValidKey(string(longKey))
	require.Nil(t, err)

	longKey = append(longKey, 32)
	err = isValidKey(string(longKey))
	require.NotNil(t, err)

	nonUTF8, _ := utf8ToGbk([]byte("非utf8测试"))
	err = isValidKey(string(nonUTF8))
	require.NotNil(t, err)

	invisiableString1 := string([]byte{0, 1, 2, 3, 4, 5})
	err = isValidKey(invisiableString1)
	require.NotNil(t, err)

}
