package tos

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type TimeoutErr struct {
	timeout bool
}

func (t *TimeoutErr) Error() string {
	return "client request time out"
}
func (t *TimeoutErr) Timeout() bool {
	return t.timeout
}

var (
	ClientTimeout error = &TimeoutErr{true}
	TosStatus500  error = &TosServerError{RequestInfo: RequestInfo{StatusCode: 500}}
	TosStatus429  error = &TosServerError{RequestInfo: RequestInfo{StatusCode: 429}}
	TosStatus200  error = &TosServerError{RequestInfo: RequestInfo{StatusCode: 200}}
)

func genWork(returns []error) (func() error, *int) {
	i := 0
	work := func() error {
		i++
		if i > len(returns) {
			return nil
		}
		return returns[i-1]
	}
	return work, &i
}

func TestStatusRetrierBase(t *testing.T) {
	tests := []struct {
		errs   []error
		expect int
	}{
		{
			[]error{},
			1,
		},
		{
			[]error{TosStatus500},
			2,
		},
		{
			[]error{TosStatus500, TosStatus429},
			3,
		},
		{
			[]error{TosStatus500, TosStatus429, ClientTimeout},
			4,
		},
		{
			[]error{TosStatus500, TosStatus429, ClientTimeout, ClientTimeout},
			4,
		},
	}
	r := newRetryer(exponentialBackoff(3, 10*time.Millisecond))

	for _, tt := range tests {
		work, count := genWork(tt.errs)
		r.Run(context.Background(), work, StatusCodeClassifier{})
		require.Equal(t, tt.expect, *count)
	}

}

func TestServerErrorRetrierBase(t *testing.T) {
	// HTTP POST method return retry if status code is 5xx
	tests := []struct {
		errs   []error
		expect int
	}{
		{
			[]error{},
			1,
		},
		{
			[]error{TosStatus500},
			2,
		},
		{
			[]error{TosStatus500, TosStatus429},
			2,
		},
		{
			[]error{TosStatus500, TosStatus429, ClientTimeout},
			2,
		},
		{
			[]error{TosStatus500, TosStatus429, ClientTimeout, ClientTimeout},
			2,
		},
	}
	r := newRetryer(exponentialBackoff(3, 10*time.Millisecond))
	for _, tt := range tests {
		work, count := genWork(tt.errs)
		r.Run(context.Background(), work, ServerErrorClassifier{})
		require.Equal(t, tt.expect, *count)
	}
}
