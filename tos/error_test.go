package tos

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnexpectedStatusCodeError(t *testing.T) {
	err := NewUnexpectedStatusCodeError(http.StatusBadRequest, http.StatusOK)
	require.NotNil(t, err)
	require.Equal(t, err.StatusCode, http.StatusBadRequest)
	require.Equal(t, len(err.ExpectedCodes), 1)
	require.Equal(t, err.ExpectedCodes[0], http.StatusOK)

	err = err.WithRequestID("xxx")
	require.NotNil(t, err)
	require.Equal(t, err.StatusCode, http.StatusBadRequest)
	require.Equal(t, len(err.ExpectedCodes), 1)
	require.Equal(t, err.ExpectedCodes[0], http.StatusOK)
	require.Equal(t, err.RequestID, "xxx")
}

func TestRequestID(t *testing.T) {
	var err error = &Error{
		StatusCode: http.StatusBadRequest,
		Code:       "InvalidArgument",
		RequestID:  "xxx",
	}
	require.Equal(t, RequestID(err), "xxx")

	err = NewUnexpectedStatusCodeError(http.StatusBadRequest, http.StatusOK).
		WithRequestID("123")
	require.Equal(t, RequestID(err), "123")

	err = &ChecksumError{RequestID: "345"}
	require.Equal(t, RequestID(err), "345")

	err = &SerializeError{RequestID: "987"}
	require.Equal(t, RequestID(err), "987")
}
