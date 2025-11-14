package tos

import (
	"bytes"
	"context"
	"net/http"
)

func (cli *ClientV2) CreateAccessPoint(ctx context.Context, input *CreateAccessPointInput) (*CreateAccessPointOutput, error) {

	if input == nil {
		return nil, InputIsNilClientError
	}
	data, contentMD5, err := marshalInput("CreateAccessPoint", input)
	if err != nil {
		return nil, err
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithHeader(HeaderContentMD5, contentMD5).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodPut, bytes.NewReader(data), "/accesspoint/"+input.AccessPointName, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := CreateAccessPointOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) GetAccessPoint(ctx context.Context, input *GetAccessPointInput) (*GetAccessPointOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/accesspoint/"+input.AccessPointName, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := GetAccessPointOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) ListAccessPoints(ctx context.Context, input *ListAccessPointsInput) (*ListAccessPointsOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithParams(*input).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/accesspoint", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := ListAccessPointsOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) DeleteAccessPoint(ctx context.Context, input *DeleteAccessPointInput) (*DeleteAccessPointOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodDelete, nil, "/accesspoint/"+input.AccessPointName, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := DeleteAccessPointOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) ListBindAcceleratorForAccessPoint(ctx context.Context, input *ListBindAcceleratorForAccessPointInput) (*ListBindAcceleratorForAccessPointOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/accesspoint/"+input.AccessPointName+"/accelerator", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := ListBindAcceleratorForAccessPointOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) ListBindAccessPointForAccelerator(ctx context.Context, input *ListBindAccessPointForAcceleratorInput) (*ListBindAccessPointForAcceleratorOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/accelerator/"+input.AcceleratorID+"/accesspoint", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := ListBindAccessPointForAcceleratorOutput{RequestInfo: res.RequestInfo()}
	if err := marshalOutput(res, &output); err != nil {
		return nil, err
	}
	return &output, nil
}

func (cli *ClientV2) BindAcceleratorWithAccessPoint(ctx context.Context, input *BindAcceleratorWithAccessPointInput) (*BindAcceleratorWithAccessPointOutput, error) {

	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodPut, nil, "/accesspoint/"+input.AccessPointName+"/accelerator/"+input.AcceleratorID, cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := BindAcceleratorWithAccessPointOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) UnbindAcceleratorWithAccessPoint(ctx context.Context, input *UnbindAcceleratorWithAccessPointInput) (*UnbindAcceleratorWithAccessPointOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}
	res, err := cli.newControlBuilder(input.AccountID).
		SetGeneric(input.GenericInput).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodDelete, nil, "/accesspoint/"+input.AccessPointName+"/accelerator/"+input.AcceleratorID, cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := UnbindAcceleratorWithAccessPointOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}
