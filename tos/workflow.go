package tos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

// CreateWorkflow 创建或更新 Bucket 的数据处理工作流配置（事件触发模式）。
func (cli *ClientV2) CreateWorkflow(ctx context.Context, input *PutConvertWorkflowInput) (*PutConvertWorkflowOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	data, contentMD5, err := marshalInput("PutConvertWorkflowInput", struct {
		Role  string         `json:"Role,omitempty"`
		Rules []WorkflowRule `json:"Rules,omitempty"`
	}{
		Role:  input.Role,
		Rules: input.Rules,
	})
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("workflow", "").
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &PutConvertWorkflowOutput{RequestInfo: res.RequestInfo()}, nil
}

// GetWorkflow 获取 Bucket 的数据处理工作流配置。
func (cli *ClientV2) GetWorkflow(ctx context.Context, input *GetConvertWorkflowInput) (*GetConvertWorkflowOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("workflow", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetConvertWorkflowOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// DeleteWorkflow 删除 Bucket 的数据处理工作流配置。
func (cli *ClientV2) DeleteWorkflow(ctx context.Context, input *DeleteConvertWorkflowInput) (*DeleteConvertWorkflowOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("workflow", "").
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &DeleteConvertWorkflowOutput{RequestInfo: res.RequestInfo()}, nil
}

// GetWorkflowExecution 查询单个工作流执行实例的详情。
func (cli *ClientV2) GetWorkflowExecution(ctx context.Context, input *GetWorkflowExecutionInput) (*GetWorkflowExecutionOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if input.ExecutionID == "" {
		return nil, fmt.Errorf("tos: ExecutionID is required")
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("workflow_execution", "").
		WithQuery("id", input.ExecutionID).
		WithRetry(nil, StatusCodeClassifier{}).
		Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := GetWorkflowExecutionOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// ListWorkflowExecution 分页查询工作流执行实例列表。
func (cli *ClientV2) ListWorkflowExecution(ctx context.Context, input *ListWorkflowExecutionInput) (*ListWorkflowExecutionOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}

	rb := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("workflow_execution", "").
		WithRetry(nil, StatusCodeClassifier{})
	if input.PageSize > 0 {
		rb.WithQuery("page_size", strconv.Itoa(input.PageSize))
	}
	if input.PageToken != "" {
		rb.WithQuery("page_token", input.PageToken)
	}
	if input.StartTime > 0 {
		rb.WithQuery("start_time", strconv.FormatInt(input.StartTime, 10))
	}
	if input.EndTime > 0 {
		rb.WithQuery("end_time", strconv.FormatInt(input.EndTime, 10))
	}

	res, err := rb.Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := ListWorkflowExecutionOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// CreateWorkflowJob 提交工作流 Job（手动触发模式）。
// 路由统一走 media_jobs query，后端根据 job_type 分发到对应 handler。
func (cli *ClientV2) CreateWorkflowJob(ctx context.Context, input *CreateWorkflowJobInput) (*CreateWorkflowJobOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if input.JobType == "" {
		return nil, fmt.Errorf("tos: JobType is required")
	}
	if input.JobDetail == nil {
		return nil, fmt.Errorf("tos: JobDetail is required")
	}

	data, _, err := marshalInput("CreateWorkflowJobInput", input.JobDetail)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("media_jobs", "").
		WithQuery("job_type", string(input.JobType)).
		WithHeader(HeaderContentType, "application/json").
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		Request(ctx, http.MethodPost, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	output := CreateWorkflowJobOutput{RequestInfo: res.RequestInfo()}
	if err = marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

// QueryWorkflowJobs 查询工作流 Job 列表或单个 Job 详情。
// 设置 JobID 时查询单个 Job；否则按 PageSize/PageToken/StartTime/EndTime 分页查询。
func (cli *ClientV2) QueryWorkflowJobs(ctx context.Context, input *QueryWorkflowJobsInput) (*QueryWorkflowJobsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if input.JobType == "" {
		return nil, fmt.Errorf("tos: JobType is required")
	}

	rb := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("job_type", input.JobType).
		WithRetry(nil, StatusCodeClassifier{})

	if input.JobID != "" {
		rb.WithQuery("job_id", input.JobID)
	} else {
		if input.PageSize > 0 {
			rb.WithQuery("page_size", strconv.Itoa(input.PageSize))
		}
		if input.PageToken != "" {
			rb.WithQuery("page_token", input.PageToken)
		}
		if input.StartTime > 0 {
			rb.WithQuery("start_time", strconv.FormatInt(input.StartTime, 10))
		}
		if input.EndTime > 0 {
			rb.WithQuery("end_time", strconv.FormatInt(input.EndTime, 10))
		}
	}

	res, err := rb.Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var raw struct {
		JobType       string            `json:"JobType,omitempty"`
		Items         []json.RawMessage `json:"Items,omitempty"`
		NextPageToken string            `json:"NextPageToken,omitempty"`
	}
	if err = marshalOutput(res, &raw); err != nil {
		return nil, err
	}

	return &QueryWorkflowJobsOutput{
		RequestInfo:   res.RequestInfo(),
		JobType:       raw.JobType,
		Items:         raw.Items,
		NextPageToken: raw.NextPageToken,
	}, nil
}

// ListProcessTemplate 列举处理模板，按 Tag 和 Category 筛选。
func (cli *ClientV2) ListProcessTemplate(ctx context.Context, input *ListProcessTemplateInput) (*ListProcessTemplateOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if input.Tag == "" {
		return nil, fmt.Errorf("tos: Tag is required")
	}

	rb := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("process_template", "").
		WithQuery("tag", input.Tag)

	if input.Category != "" {
		rb.WithQuery("category", input.Category)
	}

	res, err := rb.Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var templates []json.RawMessage
	if err = marshalOutput(res, &templates); err != nil {
		return nil, err
	}

	return &ListProcessTemplateOutput{
		RequestInfo: res.RequestInfo(),
		Templates:   templates,
	}, nil
}

// GetProcessTemplate 获取指定处理模板的详情。
func (cli *ClientV2) GetProcessTemplate(ctx context.Context, input *GetProcessTemplateInput) (*GetProcessTemplateOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if input.Tag == "" {
		return nil, fmt.Errorf("tos: Tag is required")
	}
	if input.ID == "" {
		return nil, fmt.Errorf("tos: ID is required")
	}

	rb := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("process_template", "").
		WithQuery("tag", input.Tag).
		WithQuery("id", input.ID)

	if input.Category != "" {
		rb.WithQuery("category", input.Category)
	}

	res, err := rb.Request(ctx, http.MethodGet, nil, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var template json.RawMessage
	if err = marshalOutput(res, &template); err != nil {
		return nil, err
	}

	return &GetProcessTemplateOutput{
		RequestInfo: res.RequestInfo(),
		Template:    template,
	}, nil
}

// PutProcessTemplate 创建处理模板，返回 TemplateID。
func (cli *ClientV2) PutProcessTemplate(ctx context.Context, input *PutProcessTemplateInput) (*PutProcessTemplateOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if input.Tag == "" {
		return nil, fmt.Errorf("tos: Tag is required")
	}
	if input.TemplateConfig == nil {
		return nil, fmt.Errorf("tos: TemplateConfig is required")
	}

	data, _, err := marshalInput("PutProcessTemplateInput", input.TemplateConfig)
	if err != nil {
		return nil, err
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("process_template", "").
		WithQuery("tag", input.Tag).
		WithHeader(HeaderContentType, "application/json").
		Request(ctx, http.MethodPut, bytes.NewReader(data), cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return &PutProcessTemplateOutput{
		RequestInfo: res.RequestInfo(),
		TemplateID:  strings.TrimSpace(string(body)),
	}, nil
}

// DeleteProcessTemplate 删除指定处理模板。
func (cli *ClientV2) DeleteProcessTemplate(ctx context.Context, input *DeleteProcessTemplateInput) (*DeleteProcessTemplateOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	if err := isValidBucketName(input.Bucket, cli.isCustomDomain); err != nil {
		return nil, err
	}
	if input.ID == "" {
		return nil, fmt.Errorf("tos: ID is required")
	}

	res, err := cli.newBuilder(input.Bucket, "").
		SetGeneric(input.GenericInput).
		WithQuery("process_template", "").
		WithQuery("id", input.ID).
		Request(ctx, http.MethodDelete, nil, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()

	return &DeleteProcessTemplateOutput{RequestInfo: res.RequestInfo()}, nil
}
