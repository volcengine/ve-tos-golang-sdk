package tos

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos/enum"
)

func TestPcmDataProcessOutputUnmarshal(t *testing.T) {
	data := []byte(`{
		"bucket": "data-process-bench",
		"object": "output.pcm",
		"object_size": "32176542",
		"status": "OK"
	}`)

	var output PcmDataProcessOutput
	err := json.Unmarshal(data, &output)
	require.NoError(t, err)
	require.Equal(t, "data-process-bench", output.PcmBucket)
	require.Equal(t, "output.pcm", output.PcmObject)
	require.EqualValues(t, 32176542, output.PcmObjectSize)
	require.Equal(t, enum.VideoDataProcessStatusOK, output.PcmStatus)
}

