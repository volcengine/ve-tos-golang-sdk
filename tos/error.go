package tos

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type Error struct {
	StatusCode int    `json:"-"`
	Code       string `json:"Code,omitempty"`
	Message    string `json:"Message,omitempty"`
	RequestID  string `json:"RequestId,omitempty"`
	HostID     string `json:"HostId,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("tos: request error: StatusCode=%d, Code=%s, Message=%q, RequestID=%s, HostID=%s",
		e.StatusCode, e.Code, e.Message, e.RequestID, e.HostID)
}

// Code return error code saved in tos.Error
func Code(err error) string {
	if er, ok := err.(*Error); ok {
		return er.Code
	}
	return ""
}

// StatueCode return status code saved in tos.Error or tos.UnexpectedStatusCodeError
func StatueCode(err error) int {
	if er, ok := err.(*Error); ok {
		return er.StatusCode
	}

	if er, ok := err.(*UnexpectedStatusCodeError); ok {
		return er.StatusCode
	}

	return 0
}

func RequestID(err error) string {
	switch ev := err.(type) {
	case *Error:
		return ev.RequestID
	case *UnexpectedStatusCodeError:
		return ev.RequestID
	case *ChecksumError:
		return ev.RequestID
	case *SerializeError:
		return ev.RequestID
	}
	return ""
}

type UnexpectedStatusCodeError struct {
	StatusCode    int    `json:"StatusCode,omitempty"`
	ExpectedCodes []int  `json:"ExpectedCodes,omitempty"`
	RequestID     string `json:"RequestId,omitempty"`
	expectedCodes [2]int
}

func NewUnexpectedStatusCodeError(statusCode int, expectedCode int, expectedCodes ...int) *UnexpectedStatusCodeError {
	err := UnexpectedStatusCodeError{
		StatusCode: statusCode,
	}
	err.ExpectedCodes = err.expectedCodes[:0]
	err.ExpectedCodes = append(err.ExpectedCodes, expectedCode)
	err.ExpectedCodes = append(err.ExpectedCodes, expectedCodes...)
	return &err
}

func (us *UnexpectedStatusCodeError) WithRequestID(requestID string) *UnexpectedStatusCodeError {
	us.RequestID = requestID
	return us
}

func (us *UnexpectedStatusCodeError) GoString() string {
	return fmt.Sprintf("tos.UnexpectedStatusCodeError{StatusCode:%d, ExpectedCodes:%v, RequestID:%s}",
		us.StatusCode, us.ExpectedCodes, us.RequestID)
}

func (us *UnexpectedStatusCodeError) Error() string {
	return fmt.Sprintf("tos: unexptected status code error: StatusCode=%d, ExpectedCodes=%v, RequestID=%s",
		us.StatusCode, us.ExpectedCodes, us.RequestID)
}

type ChecksumError struct {
	RequestID        string `json:"RequestId,omitempty"`
	ExpectedChecksum string `json:"ExpectedChecksum,omitempty"`
	ActualChecksum   string `json:"ActualChecksum,omitempty"`
}

func (ce *ChecksumError) Error() string {
	return fmt.Sprintf("tos: checksum error: RequestID=%s, ExpectedChecksum=%s, ActualChecksum=%s",
		ce.RequestID, ce.ExpectedChecksum, ce.ActualChecksum)
}

type SerializeError struct {
	RequestID string `json:"RequestId,omitempty"`
	Message   string `json:"Message,omitempty"`
}

func (se *SerializeError) Error() string {
	return fmt.Sprintf("tos: serialize error: RequestID=%s, Message=%q", se.RequestID, se.Message)
}

func checkError(res *Response, okCode int, okCodes ...int) error {
	if res.StatusCode == okCode {
		return nil
	}

	for _, code := range okCodes {
		if res.StatusCode == code {
			return nil
		}
	}

	defer res.Close()

	if res.StatusCode >= http.StatusBadRequest && res.Body != nil {
		data, err := ioutil.ReadAll(io.LimitReader(res.Body, 64<<10)) // avoid too large
		if err == nil && len(data) > 0 {
			se := Error{StatusCode: res.StatusCode}
			if err = json.Unmarshal(data, &se); err == nil {
				return &se
			}
		}
		// fall through
	}

	return NewUnexpectedStatusCodeError(res.StatusCode, okCode, okCodes...).
		WithRequestID(res.RequestInfo().RequestID)
}
