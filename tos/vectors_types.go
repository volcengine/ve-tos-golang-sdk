package tos

import (
	"time"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

// Request
type CreateVectorBucketInput struct  {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required
}

// Response
type CreateVectorsBucketOutput struct {
    RequestInfo
}

// Request
type GetVectorBucketInput struct  {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required
    AccountID string `json:"-"` // required
}

// Response
type GetVectorBucketOutput struct {
    RequestInfo
    VectorBucket VectorBucket `json:"vectorBucket"`
}

type VectorBucket struct {
    CreationTime time.Time `json:"-"`
    VectorBucketName string `json:"vectorBucketName"`
    VectorBucketTrn string `json:"vectorBucketTrn"`
    ProjectName string `json:"projectName"`
}

// Request
type DeleteVectorBucketInput struct  {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required
    AccountID string `json:"-"` // required
}

// Response
type DeleteVectorBucketOutput struct {
    RequestInfo
}

// Request
type ListVectorBucketsInput struct {
    GenericInput `json:"-"`
    MaxResults int    `json:"maxResults,omitempty"`
    NextToken  string `json:"nextToken,omitempty"`
    Prefix     string `json:"prefix,omitempty"`
}

// Response
type ListVectorBucketsOutput struct {
    RequestInfo
    NextToken     string                 `json:"nextToken,omitempty"`
    VectorBuckets []VectorBucketSummary `json:"vectorBuckets"`
}

type VectorBucketSummary struct {
    CreationTime     time.Time `json:"creationTime"`
    VectorBucketName string    `json:"vectorBucketName"`
    VectorBucketTrn  string    `json:"vectorBucketTrn"`
    ProjectName      string    `json:"projectName"`
}

// Request
type CreateIndexInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required
    AccountID string `json:"-"` // required
    IndexName string `json:"indexName"` // required
    DataType enum.DataType `json:"dataType,omitempty"`
    Dimension int `json:"dimension,omitempty"`
    DistanceMetric enum.DistanceMetricType `json:"distanceMetric,omitempty"`
    MetadataConfiguration MetadataConfiguration `json:"metadataConfiguration,omitempty"`
}

type MetadataConfiguration struct {
    NonFilterableMetadataKeys []string `json:"nonFilterableMetadataKeys,omitempty"`
}

// Response
type CreateIndexOutput struct {
    RequestInfo
}

// Request
type GetIndexInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required 
    AccountID string `json:"-"` // required
    IndexName string `json:"indexName"` // required
}

// Response
type GetIndexOutput struct {
    RequestInfo
    Index Index `json:"index"`
}

type Index struct {
    CreationTime time.Time `json:"-"`
    DataType enum.DataType `json:"dataType"`
    Dimension int `json:"dimension"`
    DistanceMetric enum.DistanceMetricType `json:"distanceMetric"`
    MetadataConfiguration MetadataConfiguration `json:"metadataConfiguration"`
    IndexName string `json:"indexName"`
    IndexTrn string `json:"indexTrn"`
    VectorBucketName string `json:"vectorBucketName"`
    CreationTimeRaw int64 `json:"creationTime"`
}

// Request
type DeleteIndexInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required 
    AccountID string `json:"-"` // required
    IndexName string `json:"indexName"` // required
}

// Response
type DeleteIndexOutput struct {
    RequestInfo
}

// Request
type ListIndexesInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required
    AccountID string `json:"-"` // required
    MaxResults int `json:"maxResults,omitempty"`
    NextToken string `json:"nextToken,omitempty"`
    Prefix string `json:"prefix,omitempty"`
}

// Response
type ListIndexesOutput struct {
    RequestInfo
    NextToken string `json:"nextToken,omitempty"`
    Indexes []IndexSummary `json:"indexes"`
}

type IndexSummary struct {
    CreationTime time.Time `json:"-"`
    IndexName string `json:"indexName"`
    IndexTrn string `json:"indexTrn"`
    VectorBucketName string `json:"vectorBucketName"`
    CreationTimeRaw int64 `json:"creationTime"`
}

// Request
type PutVectorsInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required 
    AccountID string `json:"-"` // required
    IndexName string `json:"indexName"` // required
    Vectors []Vector `json:"vectors"`
}

type Vector struct {
    Key string `json:"key"`
    Data VectorData `json:"data"`
    Metadata map[string]interface{} `json:"metadata,omitempty"` // json format
}

type VectorData struct {
    Value []float32 `json:"float32"`
}

// Response
type PutVectorsOutput struct {
    RequestInfo
}

// Request
type GetVectorsInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required 
    AccountID string `json:"-"` // required
    IndexName string `json:"indexName"` // required
    Keys []string `json:"keys"` // required
    ReturnData bool `json:"returnData,omitempty"`
    ReturnMetadata bool `json:"returnMetadata,omitempty"`
}

// Response
type GetVectorsOutput struct {
    RequestInfo
    Vectors []Vector `json:"vectors"`
}

// Request
type DeleteVectorsInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required 
    AccountID string `json:"-"` // required
    IndexName string `json:"indexName"` // required
    Keys []string `json:"keys"` // required
}

// Response
type DeleteVectorsOutput struct {
    RequestInfo
}

// Request
type QueryVectorsInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required 
    AccountID string `json:"-"` // required
    IndexName string `json:"indexName"` // required
    ReturnDistance bool `json:"returnDistance,omitempty"`
    ReturnMetadata bool `json:"returnMetadata,omitempty"`
    TopK int `json:"topK"`
    QueryVector VectorData `json:"queryVector"`
    Filter map[string]interface{} `json:"filter,omitempty"` // json format
}

// Response
type QueryVectorsOutput struct {
    RequestInfo
    Vectors []DistanceVector `json:"vectors"`
}

type DistanceVector struct {
    Key string `json:"key"`
    Data VectorData `json:"data"`
    Distance float32 `json:"distance,omitempty"`
    Metadata map[string]interface{} `json:"metadata,omitempty"` // json format
}

// Request
type ListVectorsInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required 
    AccountID string `json:"-"` // required
    IndexName string `json:"indexName"` // required
    MaxResults int `json:"maxResults,omitempty"`
    NextToken string `json:"nextToken,omitempty"`
    ReturnData bool `json:"returnData,omitempty"`
    ReturnMetadata bool `json:"returnMetadata,omitempty"`
}

// Response
type ListVectorsOutput struct {
    RequestInfo
    NextToken string `json:"nextToken,omitempty"`
    Vectors []Vector `json:"vectors"`
}

// Request
type PutVectorBucketPolicyInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required
    AccountID string `json:"-"` // required
    Policy string `json:"policy"` // required
}

// Response
type PutVectorBucketPolicyOutput struct {
    RequestInfo
}

// Request
type GetVectorBucketPolicyInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required
    AccountID string `json:"-"` // required
}

// Response
type GetVectorBucketPolicyOutput struct {
    RequestInfo
    Policy string `json:"policy"`
}

// Request
type DeleteVectorBucketPolicyInput struct {
    GenericInput `json:"-"`
    VectorBucketName string `json:"vectorBucketName"` // required
    AccountID string `json:"-"` // required
}

// Response
type DeleteVectorBucketPolicyOutput struct {
    RequestInfo
}