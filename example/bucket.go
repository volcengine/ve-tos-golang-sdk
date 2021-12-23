package main

import (
	"context"
	"fmt"
	"os"

	"github.com/volcengine/ve-tos-golang-sdk/tos"
)

var (
	endpoint  = os.Getenv("TOS_GO_SDK_ENDPOINT")
	region    = os.Getenv("TOS_GO_SDK_REGION")
	accessKey = os.Getenv("TOS_GO_SDK_AK")
	secretKey = os.Getenv("TOS_GO_SDK_SK")
	//bucket    = os.Getenv("TOS_GO_SDK_BUCKET")
)

func check(err error, message string) {
	if err != nil {
		fmt.Printf("%s err: %s\n", message, err.Error())
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) != 0 {
		fmt.Println("Usage: ./bucket bucket-name")
		os.Exit(1)
	}

	bucket := os.Args[1]
	client, err := tos.NewClient(endpoint, tos.WithRegion(region),
		tos.WithCredentials(tos.NewStaticCredentials(accessKey, secretKey)))
	check(err, "NewClient")

	created, err := client.CreateBucket(context.Background(), &tos.CreateBucketInput{Bucket: bucket})
	check(err, "CreateBucket")
	fmt.Printf("bucket created: %+v\n", created)

	buckets, err := client.ListBuckets(context.Background(), &tos.ListBucketsInput{})
	check(err, "ListBuckets")
	fmt.Printf("listed buckets: %+v\n", buckets)

	head, err := client.HeadBucket(context.Background(), bucket)
	check(err, "HeadBucket")
	fmt.Printf("head buckets: %+v\n", head)

	deleted, err := client.DeleteBucket(context.Background(), bucket)
	check(err, "DeleteBucket")
	fmt.Printf("delete buckets: %+v\n", deleted)
}
