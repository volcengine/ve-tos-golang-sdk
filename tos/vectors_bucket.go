package tos

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"time"
)

type vectorBucket struct {
	CreationTime     int64  `json:"creationTime"`
	VectorBucketName string `json:"vectorBucketName"`
	VectorBucketTrn  string `json:"vectorBucketTrn"`
	ProjectName      string `json:"projectName"`
}

type getVectorBucketOutput struct {
	RequestInfo
	VectorBucket vectorBucket `json:"vectorBucket"`
}

type listVectorBucketsOutput struct {
	RequestInfo
	NextToken     string                `json:"nextToken,omitempty"`
	VectorBuckets []vectorBucketSummary `json:"vectorBuckets"`
}

type vectorBucketSummary struct {
	CreationTime     int64  `json:"creationTime"`
	VectorBucketName string `json:"vectorBucketName"`
	VectorBucketTrn  string `json:"vectorBucketTrn"`
	ProjectName      string `json:"projectName"`
}

func (client *TosVectorsClient) CreateVectorBucket(ctx context.Context, input *CreateVectorBucketInput) (*CreateVectorsBucketOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("CreateVectorBucketInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder("", "").
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/CreateVectorBucket", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &CreateVectorsBucketOutput{
		RequestInfo: res.RequestInfo(),
	}, nil
}

func (client *TosVectorsClient) GetVectorBucket(ctx context.Context, input *GetVectorBucketInput) (*GetVectorBucketOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("GetVectorBucketInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/GetVectorBucket", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	getOutput := &getVectorBucketOutput{}
	if err = marshalOutput(res, getOutput); err != nil {
		return nil, err
	}
	output := &GetVectorBucketOutput{RequestInfo: res.RequestInfo()}
	output.VectorBucket.CreationTime = time.Unix(int64(getOutput.VectorBucket.CreationTime), 0)
	output.VectorBucket.VectorBucketName = getOutput.VectorBucket.VectorBucketName
	output.VectorBucket.VectorBucketTrn = getOutput.VectorBucket.VectorBucketTrn
	output.VectorBucket.ProjectName = getOutput.VectorBucket.ProjectName
	return output, nil
}

func (client *TosVectorsClient) DeleteVectorBucket(ctx context.Context, input *DeleteVectorBucketInput) (*DeleteVectorBucketOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("DeleteVectorBucketInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/DeleteVectorBucket", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &DeleteVectorBucketOutput{
		RequestInfo: res.RequestInfo(),
	}, nil
}

func (client *TosVectorsClient) ListVectorBuckets(ctx context.Context, input *ListVectorBucketsInput) (*ListVectorBucketsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	data, _, err := marshalInput("ListVectorBucketsInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder("", "").
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/ListVectorBuckets", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	listOutput := &listVectorBucketsOutput{}

	if err = marshalOutput(res, &listOutput); err != nil {
		return nil, err
	}
	output := &ListVectorBucketsOutput{RequestInfo: res.RequestInfo()}
	output.NextToken = listOutput.NextToken
	for _, bucket := range listOutput.VectorBuckets {
		output.VectorBuckets = append(output.VectorBuckets, VectorBucketSummary{
			CreationTime:     time.Unix(int64(bucket.CreationTime), 0),
			VectorBucketName: bucket.VectorBucketName,
			VectorBucketTrn:  bucket.VectorBucketTrn,
			ProjectName:      bucket.ProjectName,
		})
	}
	return output, nil
}

func (client *TosVectorsClient) PutVectorBucketPolicy(ctx context.Context, input *PutVectorBucketPolicyInput) (*PutVectorBucketPolicyOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if input.Policy == "" {
		return nil, newTosClientError("tosvectors: Policy is empty", nil)
	}

	data, _, err := marshalInput("PutVectorBucketPolicyInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/PutVectorBucketPolicy", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &PutVectorBucketPolicyOutput{
		RequestInfo: res.RequestInfo(),
	}, nil
}

func (client *TosVectorsClient) GetVectorBucketPolicy(ctx context.Context, input *GetVectorBucketPolicyInput) (*GetVectorBucketPolicyOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("GetVectorBucketPolicyInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/GetVectorBucketPolicy", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	policyData, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	output := &GetVectorBucketPolicyOutput{RequestInfo: res.RequestInfo(), Policy: string(policyData)}

	return output, nil
}

func (client *TosVectorsClient) DeleteVectorBucketPolicy(ctx context.Context, input *DeleteVectorBucketPolicyInput) (*DeleteVectorBucketPolicyOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("DeleteVectorBucketPolicyInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/DeleteVectorBucketPolicy", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &DeleteVectorBucketPolicyOutput{
		RequestInfo: res.RequestInfo(),
	}, nil
}
