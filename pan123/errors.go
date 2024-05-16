package pan123

import (
	"fmt"
)

const (
	defaultTraceID = "no_trace_id"
)

type SDKError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

func (e *SDKError) Error() string {
	return fmt.Sprintf("SDKError: [%d](%s): %s", e.Code, e.TraceID, e.Message)
}

func newSDKError(code int, message, traceID string) error {
	sdkErr := new(SDKError)
	sdkErr.Code = code
	sdkErr.Message = message
	sdkErr.TraceID = traceID
	return sdkErr
}
