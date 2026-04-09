package tos

import (
	"bytes"
	"context"
	"net/http"
)

// QosConfig defines QoS configuration for bucket ObjectSet.
type QosConfig struct {
	ReadsQps   int `json:"ReadsQps,omitempty"`
	WritesQps  int `json:"WritesQps,omitempty"`
	ListQps    int `json:"ListQps,omitempty"`
	ReadsRate  int `json:"ReadsRate,omitempty"`
	WritesRate int `json:"WritesRate,omitempty"`
}

// PutBucketObjectSetConfigurationInput is the input for PutBucketObjectSetConfiguration.
type PutBucketObjectSetConfigurationInput struct {
	Bucket                 string    `json:"-"`
	PathLevel              int       `json:"PathLevel"`
	CustomDelimiter        string    `json:"CustomDelimiter,omitempty"`
	EnableDefaultObjectSet bool      `json:"EnableDefaultObjectSet"`
	StorageQuota           string    `json:"StorageQuota,omitempty"`
	Qos                    QosConfig `json:"Qos,omitempty"`
	GenericInput           `json:"-"`
}

// PutBucketObjectSetConfigurationOutput is the output for PutBucketObjectSetConfiguration.
type PutBucketObjectSetConfigurationOutput struct {
	RequestInfo
}

// GetBucketObjectSetConfigurationInput is the input for GetBucketObjectSetConfiguration.
type GetBucketObjectSetConfigurationInput struct {
	Bucket       string `json:"-"`
	GenericInput `json:"-"`
}

// GetBucketObjectSetConfigurationOutput is the output for GetBucketObjectSetConfiguration.
type GetBucketObjectSetConfigurationOutput struct {
	RequestInfo
	PathLevel              int       `json:"PathLevel"`
	CustomDelimiter        string    `json:"CustomDelimiter"`
	EnableDefaultObjectSet bool      `json:"EnableDefaultObjectSet"`
	Qos                    QosConfig `json:"Qos"`
	StorageQuota           string    `json:"StorageQuota"`
}

// PutBucketObjectSetConfiguration sets ObjectSet configuration on a bucket.
func (cli *ClientV2) PutBucketObjectSetConfiguration(ctx context.Context,
	input *PutBucketObjectSetConfigurationInput) (*PutBucketObjectSetConfigurationOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("PutBucketObjectSetConfiguration", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsetconfiguration", "").
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := PutBucketObjectSetConfigurationOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

// GetBucketObjectSetConfiguration gets ObjectSet configuration of a bucket.
func (cli *ClientV2) GetBucketObjectSetConfiguration(ctx context.Context,
	input *GetBucketObjectSetConfigurationInput) (*GetBucketObjectSetConfigurationOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsetconfiguration", "").
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetBucketObjectSetConfigurationOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// PutObjectSetQuotaByTag sets ObjectSet quota rules based on tags for a bucket.
func (cli *ClientV2) PutObjectSetQuotaByTag(ctx context.Context,
	input *PutObjectSetQuotaByTagInput) (*PutObjectSetQuotaByTagOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("PutObjectSetQuotaByTag", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsetquotabytag", "").
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := PutObjectSetQuotaByTagOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

// GetObjectSetQuotaByTag gets ObjectSet quota rules based on tags for a bucket.
func (cli *ClientV2) GetObjectSetQuotaByTag(ctx context.Context,
	input *GetObjectSetQuotaByTagInput) (*GetObjectSetQuotaByTagOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsetquotabytag", "").
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetObjectSetQuotaByTagOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// DeleteObjectSetQuotaByTag deletes ObjectSet quota rules based on tags for a bucket.
func (cli *ClientV2) DeleteObjectSetQuotaByTag(ctx context.Context,
	input *DeleteObjectSetQuotaByTagInput) (*DeleteObjectSetQuotaByTagOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsetquotabytag", "").
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := DeleteObjectSetQuotaByTagOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

// PutObjectSetQuota sets storage quota for a specified object set.
func (cli *ClientV2) PutObjectSetQuota(ctx context.Context, input *PutObjectSetQuotaInput) (*PutObjectSetQuotaOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("PutObjectSetQuota", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsetquota", "").
		WithParams(*input).
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := PutObjectSetQuotaOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

// GetObjectSetQuota gets storage quota of a specified object set.
func (cli *ClientV2) GetObjectSetQuota(ctx context.Context, input *GetObjectSetQuotaInput) (*GetObjectSetQuotaOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsetquota", "").
		WithParams(*input).
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetObjectSetQuotaOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// GetObjectSetStorage gets storage usage statistics of a specified object set.
func (cli *ClientV2) GetObjectSetStorage(ctx context.Context, input *GetObjectSetStorageInput) (*GetObjectSetStorageOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsetstorage", "").
		WithParams(*input).
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetObjectSetStorageOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// PutObjectSet creates or updates an object set with specified tags.
func (cli *ClientV2) PutObjectSet(ctx context.Context, input *PutObjectSetInput) (*PutObjectSetOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("PutObjectSetInput", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectset", "").
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := PutObjectSetOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

// DeleteObjectSet deletes a specified object set.
func (cli *ClientV2) DeleteObjectSet(ctx context.Context, input *DeleteObjectSetInput) (*DeleteObjectSetOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectset", "").
		WithParams(*input).
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := DeleteObjectSetOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

// ListObjectSet lists object sets in a bucket, with optional filtering.
func (cli *ClientV2) ListObjectSet(ctx context.Context, input *ListObjectSetInput) (*ListObjectSetOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsets", "").
		WithParams(*input).
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListObjectSetOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// GetObjectSet retrieves the tags of a specified object set.
func (cli *ClientV2) GetObjectSet(ctx context.Context, input *GetObjectSetInput) (*GetObjectSetOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectset", "").
		WithParams(*input).
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetObjectSetOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// PutObjectSetTagging sets tags for a specified object set.
func (cli *ClientV2) PutObjectSetTagging(ctx context.Context,
	input *PutObjectSetTaggingInput) (*PutObjectSetTaggingOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, _, err := marshalInput("PutObjectSetTagging", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsettagging", "").
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := PutObjectSetTaggingOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

// GetObjectSetTagging retrieves tags of a specified object set.
func (cli *ClientV2) GetObjectSetTagging(ctx context.Context,
	input *GetObjectSetInput) (*GetObjectSetOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("objectsettagging", "").
		WithParams(*input).
		WithRetry(OnRetryFromStart, ServerErrorClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetObjectSetOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}
