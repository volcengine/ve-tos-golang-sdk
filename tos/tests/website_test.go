package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestBucketWebsite(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("website")
		client = env.prepareClient(bucket)
		ctx    = context.Background()
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	redirectAllRequestsTo := &tos.RedirectAllRequestsTo{
		HostName: "www.volcengine.com",
		Protocol: "https",
	}
	indexDocument := &tos.IndexDocument{
		Suffix:          "index.html",
		ForbiddenSubDir: true,
	}
	errDocument := &tos.ErrorDocument{Key: "err.html"}
	rule1 := tos.RoutingRule{
		Condition: tos.RoutingRuleCondition{
			KeyPrefixEquals:             "website",
			HttpErrorCodeReturnedEquals: 404,
		},
		Redirect: tos.RoutingRuleRedirect{
			Protocol:             enum.ProtocolHttp,
			HostName:             "www.volcengine.com",
			ReplaceKeyPrefixWith: "voc",
			HttpRedirectCode:     301,
		},
	}
	rule2 := tos.RoutingRule{
		Condition: tos.RoutingRuleCondition{
			KeyPrefixEquals:             "website",
			HttpErrorCodeReturnedEquals: 404,
		},
		Redirect: tos.RoutingRuleRedirect{
			Protocol:         enum.ProtocolHttps,
			HostName:         "www.volcengine.com",
			ReplaceKeyWith:   "voc",
			HttpRedirectCode: 301,
		},
	}
	routingRule := &tos.RoutingRules{Rules: []tos.RoutingRule{rule1, rule2}}
	input := &tos.PutBucketWebsiteInput{
		Bucket:                bucket,
		RedirectAllRequestsTo: redirectAllRequestsTo,
	}
	putRes, err := client.PutBucketWebsite(ctx, input)
	require.Nil(t, err)
	t.Log(putRes)

	getRes, err := client.GetBucketWebsite(ctx, &tos.GetBucketWebsiteInput{Bucket: bucket})
	require.Nil(t, err)
	require.Equal(t, getRes.RedirectAllRequestsTo, redirectAllRequestsTo)
	require.Nil(t, getRes.ErrorDocument)
	require.Nil(t, getRes.IndexDocument)
	require.Nil(t, getRes.RoutingRules)

	input = &tos.PutBucketWebsiteInput{
		Bucket:        bucket,
		IndexDocument: indexDocument,
		ErrorDocument: errDocument,
		RoutingRules:  routingRule,
	}
	putRes, err = client.PutBucketWebsite(ctx, input)
	require.Nil(t, err)
	t.Log(putRes)

	getRes, err = client.GetBucketWebsite(ctx, &tos.GetBucketWebsiteInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	require.Nil(t, getRes.RedirectAllRequestsTo)
	require.Equal(t, getRes.ErrorDocument, errDocument)
	require.Equal(t, getRes.IndexDocument, indexDocument)
	require.Equal(t, len(getRes.RoutingRules), 2)
	require.Equal(t, getRes.RoutingRules[0], rule1)
	require.Equal(t, getRes.RoutingRules[1], rule2)

	deleteBucket, err := client.DeleteBucketWebsite(ctx, &tos.DeleteBucketWebsiteInput{
		Bucket: bucket,
	})
	require.Nil(t, err)
	t.Log(deleteBucket.RequestID)

	_, err = client.GetBucketWebsite(ctx, &tos.GetBucketWebsiteInput{Bucket: bucket})
	require.NotNil(t, err)

}
