package tests

import (
	"context"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
)

type testEnv struct {
	endpoint      string
	region        string
	endpoint2     string
	region2       string
	accessKey     string
	secretKey     string
	cloudFunction string
	accountId     string
	mqInstanceId  string
	mqRoleName    string
	mqAccessKeyID string
	callbackUrl   string
	t             *testing.T
}

func newTestEnv(t *testing.T) *testEnv {
	return &testEnv{
		endpoint:      os.Getenv("TOS_GO_SDK_ENDPOINT"),
		region:        os.Getenv("TOS_GO_SDK_REGION"),
		accessKey:     os.Getenv("TOS_GO_SDK_AK"),
		secretKey:     os.Getenv("TOS_GO_SDK_SK"),
		endpoint2:     os.Getenv("TOS_GO_SDK_ENDPOINT2"),
		region2:       os.Getenv("TOS_GO_SDK_REGION2"),
		cloudFunction: os.Getenv("TOS_GO_SDK_CLOUD_FUNCTION"),
		mqInstanceId:  os.Getenv("TOS_GO_SDK_MQ_INSTANCE_ID"),
		accountId:     os.Getenv("TOS_GO_SDK_ACCOUNT_ID"),
		mqRoleName:    os.Getenv("TOS_GO_SDK_MQ_ROLE_NAME"),
		mqAccessKeyID: os.Getenv("TOS_GO_SDK_MQ_ACCESSKEY_ID"),
		callbackUrl:   os.Getenv("TOS_GO_SDK_CALLBACK_URL"),
		t:             t,
	}
}

func (e testEnv) prepareClient(bucketName string, extraOptions ...tos.ClientOption) *tos.ClientV2 {
	log := logrus.New()
	log.Level = logrus.DebugLevel
	log.Formatter = &logrus.TextFormatter{DisableQuote: true}
	options := []tos.ClientOption{
		tos.WithRegion(e.region),
		tos.WithCredentials(tos.NewStaticCredentials(e.accessKey, e.secretKey)),
		tos.WithEnableVerifySSL(false),
		tos.WithLogger(log),
		tos.WithMaxRetryCount(5),
	}
	options = append(options, extraOptions...)
	client, err := tos.NewClientV2(e.endpoint, options...)
	require.Nil(e.t, err)
	if bucketName != "" {
		create, err := client.CreateBucketV2(context.Background(), &tos.CreateBucketV2Input{
			Bucket: bucketName,
		})
		checkSuccess(e.t, create, err, 200)
	}
	return client
}
