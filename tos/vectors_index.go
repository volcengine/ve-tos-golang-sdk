package tos

import (
	"bytes"
	"context"
	"net/http"
	"time"
	
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

type index struct {
	CreationTime          int64                    `json:"creationTime"`
	DataType              string                   `json:"dataType"`
	Dimension             int                      `json:"dimension"`
	DistanceMetric        string                   `json:"distanceMetric"`
	MetadataConfiguration metadataConfiguration    `json:"metadataConfiguration"`
	IndexName             string                   `json:"indexName"`
	IndexTrn              string                   `json:"indexTrn"`
	VectorBucketName      string                   `json:"vectorBucketName"`
}

type metadataConfiguration struct {
	NonFilterableMetadataKeys []string `json:"nonFilterableMetadataKeys,omitempty"`
}

type getIndexOutput struct {
	RequestInfo
	Index index `json:"index"`
}

type listIndexesOutput struct {
	RequestInfo
	NextToken string         `json:"nextToken,omitempty"`
	Indexes   []indexSummary `json:"indexes"`
}

type indexSummary struct {
	CreationTime     int64  `json:"creationTime"`
	IndexName        string `json:"indexName"`
	IndexTrn         string `json:"indexTrn"`
	VectorBucketName string `json:"vectorBucketName"`
}

func (client *TosVectorsClient) CreateIndex(ctx context.Context, input *CreateIndexInput) (*CreateIndexOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if len(input.IndexName) == 0 {
		return nil, newTosClientError("tos: missing IndexName", nil)
	}

	data, _, err := marshalInput("CreateIndexInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/CreateIndex", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &CreateIndexOutput{
		RequestInfo: res.RequestInfo(),
	}, nil
}

func (client *TosVectorsClient) GetIndex(ctx context.Context, input *GetIndexInput) (*GetIndexOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if len(input.IndexName) == 0 {
		return nil, newTosClientError("tos: missing IndexName", nil)
	}

	data, _, err := marshalInput("GetIndexInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/GetIndex", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	getOutput := &getIndexOutput{}
	if err = marshalOutput(res, getOutput); err != nil {
		return nil, err
	}
	output := &GetIndexOutput{RequestInfo: res.RequestInfo()}
	output.Index.CreationTime = time.Unix(getOutput.Index.CreationTime, 0)
	output.Index.DataType = enum.DataType(getOutput.Index.DataType)
	output.Index.Dimension = getOutput.Index.Dimension
	output.Index.DistanceMetric = enum.DistanceMetricType(getOutput.Index.DistanceMetric)
	output.Index.IndexName = getOutput.Index.IndexName
	output.Index.IndexTrn = getOutput.Index.IndexTrn
	output.Index.VectorBucketName = getOutput.Index.VectorBucketName
	
	if len(getOutput.Index.MetadataConfiguration.NonFilterableMetadataKeys) > 0 {
		output.Index.MetadataConfiguration.NonFilterableMetadataKeys = getOutput.Index.MetadataConfiguration.NonFilterableMetadataKeys
	}
	
	return output, nil
}

func (client *TosVectorsClient) DeleteIndex(ctx context.Context, input *DeleteIndexInput) (*DeleteIndexOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	if len(input.IndexName) == 0 {
		return nil, newTosClientError("tos: missing IndexName", nil)
	}

	data, _, err := marshalInput("DeleteIndexInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/DeleteIndex", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	return &DeleteIndexOutput{
		RequestInfo: res.RequestInfo(),
	}, nil
}

func (client *TosVectorsClient) ListIndexes(ctx context.Context, input *ListIndexesInput) (*ListIndexesOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	if err := IsValidVectorsBucketName(input.VectorBucketName); err != nil {
		return nil, err
	}

	if err := IsValidAccountID(input.AccountID); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("ListIndexesInput", input)
	if err != nil {
		return nil, err
	}

	res, err := client.newBuilder(input.AccountID, input.VectorBucketName).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestVectors(ctx, http.MethodPost, bytes.NewReader(data), "/ListIndexes", client.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	listOutput := &listIndexesOutput{}
	if err = marshalOutput(res, listOutput); err != nil {
		return nil, err
	}
	output := &ListIndexesOutput{
		RequestInfo: res.RequestInfo(),
		NextToken:   listOutput.NextToken,
		Indexes:     make([]IndexSummary, 0, len(listOutput.Indexes)),
	}

	for _, idx := range listOutput.Indexes {
		indexSummary := IndexSummary{
			CreationTime:     time.Unix(idx.CreationTime, 0),
			IndexName:        idx.IndexName,
			IndexTrn:         idx.IndexTrn,
			VectorBucketName: idx.VectorBucketName,
			CreationTimeRaw:  idx.CreationTime,
		}
		output.Indexes = append(output.Indexes, indexSummary)
	}

	return output, nil
}