package tests

import (
    "context"
    "os"
    "strings"
    "testing"

    "github.com/stretchr/testify/require"
    "github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

func TestPutGetObjectWithStaticCredentialsProvider(t *testing.T) {
    env := newTestEnv(t)
    bucket := generateBucketName("provider-static")
    key := "key-provider-static"

    provider := tos.NewStaticCredentialsProvider(env.accessKey, env.secretKey, "")
    client := env.prepareClientWithProvider(bucket, provider)
    defer func() {
        cleanBucket(t, client, bucket)
    }()

    content := randomString(4 * 1024)
    put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
        PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
        Content:             strings.NewReader(content),
    })
    checkSuccess(t, put, err, 200)

    get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{Bucket: bucket, Key: key})
    require.Nil(t, err)
    require.NotNil(t, get)
    require.Equal(t, 200, get.StatusCode)
}

func TestPutGetObjectWithEnvCredentialsProvider(t *testing.T) {
    env := newTestEnv(t)
    bucket := generateBucketName("provider-env")
    key := "key-provider-env"

    // backup current env
    oldAK := os.Getenv("TOS_ACCESS_KEY")
    oldSK := os.Getenv("TOS_SECRET_KEY")
    oldToken := os.Getenv("TOS_SECURITY_TOKEN")
    defer func() {
        _ = os.Setenv("TOS_ACCESS_KEY", oldAK)
        _ = os.Setenv("TOS_SECRET_KEY", oldSK)
        _ = os.Setenv("TOS_SECURITY_TOKEN", oldToken)
    }()

    // set env for provider
    _ = os.Setenv("TOS_ACCESS_KEY", env.accessKey)
    _ = os.Setenv("TOS_SECRET_KEY", env.secretKey)
    _ = os.Setenv("TOS_SECURITY_TOKEN", "")

    provider := &tos.EnvCredentialsProvider{}
    client := env.prepareClientWithProvider(bucket, provider)
    defer func() {
        cleanBucket(t, client, bucket)
    }()

    content := randomString(4 * 1024)
    put, err := client.PutObjectV2(context.Background(), &tos.PutObjectV2Input{
        PutObjectBasicInput: tos.PutObjectBasicInput{Bucket: bucket, Key: key},
        Content:             strings.NewReader(content),
    })
    checkSuccess(t, put, err, 200)

    get, err := client.GetObjectV2(context.Background(), &tos.GetObjectV2Input{Bucket: bucket, Key: key})
    require.Nil(t, err)
    require.NotNil(t, get)
    require.Equal(t, 200, get.StatusCode)
}

