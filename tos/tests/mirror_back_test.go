package tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestBucketMirrorBack(t *testing.T) {
	var (
		env    = newTestEnv(t)
		bucket = generateBucketName("mirror-back")
		client = env.prepareClient(bucket)
	)
	defer func() {
		cleanBucket(t, client, bucket)
	}()
	ctx := context.Background()
	withQuery := true
	condition := tos.Condition{
		HttpCode:   http.StatusNotFound,
		KeyPrefix:  "prefix-",
		KeySuffix:  "-suffix",
		HttpMethod: []string{http.MethodGet, http.MethodHead},
	}
	transform := tos.Transform{
		WithKeyPrefix: "prefix-",
		WithKeySuffix: "-suffix",
		ReplaceKeyPrefix: tos.ReplaceKeyPrefix{
			KeyPrefix:   "prefix-",
			ReplaceWith: "replace-",
		},
	}
	redirect := tos.Redirect{
		RedirectType:          enum.RedirectTypeMirror,
		FetchSourceOnRedirect: true,
		PassQuery:             true,
		FollowRedirect:        true,
		MirrorHeader: tos.MirrorHeader{
			PassAll: true,
			Pass:    []string{"aa", "bb"},
			Remove:  []string{"xx"},
			Set: []tos.MirrorHeaderKeyValue{{
				Key:   "kk",
				Value: "vv",
			}},
		},
		Transform: transform,
		PublicSource: tos.PublicSource{
			SourceEndpoint: tos.SourceEndpoint{
				Primary:  []string{"http://www.volcengine.com/obj/tostest/"},
				Follower: []string{"http://www.volcengine.com/obj/tostest/"},
			},
		},
		FetchHeaderToMetaDataRules: []tos.FetchHeaderToMetaDataRule{{
			SourceHeader:   "x-source-header",
			MetaDataSuffix: "meta-data-suffix",
		}},
		FetchSourceOnRedirectWithQuery: &withQuery,
	}
	putRes, err := client.PutBucketMirrorBack(ctx, &tos.PutBucketMirrorBackInput{
		Bucket: bucket,
		Rules: []tos.MirrorBackRule{{
			ID:        "1",
			Condition: condition,
			Redirect:  redirect,
		}},
	})
	require.Nil(t, err)
	require.NotNil(t, putRes)

	getRes, err := client.GetBucketMirrorBack(ctx, &tos.GetBucketMirrorBackInput{Bucket: bucket})
	require.Nil(t, err)
	require.True(t, len(getRes.Rules) == 1)
	require.Equal(t, getRes.Rules[0].Redirect, redirect)
	require.Equal(t, getRes.Rules[0].Redirect.Transform, transform)
	require.Equal(t, getRes.Rules[0].Condition, condition)
	require.Equal(t, len(getRes.Rules[0].Redirect.FetchHeaderToMetaDataRules), 1)
	require.True(t, getRes.Rules[0].Redirect.FetchSourceOnRedirectWithQuery != nil)
	require.True(t, *getRes.Rules[0].Redirect.FetchSourceOnRedirectWithQuery, withQuery)

	deleteRes, err := client.DeleteBucketMirrorBack(ctx, &tos.DeleteBucketMirrorBackInput{Bucket: bucket})
	require.Nil(t, err)
	require.NotNil(t, deleteRes)

	getRes, err = client.GetBucketMirrorBack(ctx, &tos.GetBucketMirrorBackInput{Bucket: bucket})
	require.NotNil(t, err)

}
