package tos

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
)

func (cli *ClientV2) PutQosPolicy(ctx context.Context, input *PutQosPolicyInput) (*PutQosPolicyOutput, error) {

	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodPut, strings.NewReader(input.Policy), "/qospolicy", cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := PutQosPolicyOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}

func (cli *ClientV2) GetQosPolicy(ctx context.Context, input *GetQosPolicyInput) (*GetQosPolicyOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodGet, nil, "/qospolicy", cli.roundTripper(http.StatusOK))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	qosPolicy, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	output := GetQosPolicyOutput{RequestInfo: res.RequestInfo(), Policy: string(qosPolicy)}
	return &output, nil
}

func (cli *ClientV2) DeleteQosPolicy(ctx context.Context, input *DeleteQosPolicyInput) (*DeleteQosPolicyOutput, error) {
	if input == nil {
		return nil, InputIsNilClientError
	}

	res, err := cli.newControlBuilder(input.AccountID).
		WithRetry(OnRetryFromStart, StatusCodeClassifier{}).
		RequestControl(ctx, http.MethodDelete, nil, "/qospolicy", cli.roundTripper(http.StatusNoContent))
	if err != nil {
		return nil, err
	}
	defer res.Close()
	output := DeleteQosPolicyOutput{RequestInfo: res.RequestInfo()}
	return &output, nil
}
