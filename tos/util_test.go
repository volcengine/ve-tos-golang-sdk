package tos

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWrapLimiterReader(t *testing.T) {
	data := make([]byte, 1024*1024*10)
	_, err := rand.Reader.Read(data)
	require.Nil(t, err)
	reader := bytes.NewReader(data)
	lmReader := newWrapLimiterReader(reader, int64(1))
	readData, err := io.ReadAll(lmReader)
	require.Nil(t, err)
	require.Equal(t, readData[:1], data[:1])
}

func TestWrapLimiterReader_Seek(t *testing.T) {
	data := make([]byte, 100)
	_, err := rand.Reader.Read(data)
	require.Nil(t, err)
	reader := bytes.NewReader(data)
	lmReader := newWrapLimiterReader(reader, int64(len(data)))
	// read 10 bytes
	readBuf := make([]byte, 10)
	n, err := lmReader.Read(readBuf)
	require.Nil(t, err)
	require.Equal(t, 10, n)
	require.Equal(t, int64(90), lmReader.N)

	// seek from start
	_, err = lmReader.Seek(0, io.SeekStart)
	require.Nil(t, err)
	require.Equal(t, int64(100), lmReader.N)
	n, err = lmReader.Read(readBuf)
	require.Nil(t, err)
	require.Equal(t, 10, n)
	require.Equal(t, data[:10], readBuf)

	// seek from current
	_, err = lmReader.Seek(10, io.SeekStart)
	require.Nil(t, err)
	_, err = lmReader.Seek(10, io.SeekCurrent)
	require.Nil(t, err)
	n, err = lmReader.Read(readBuf)
	require.Nil(t, err)
	require.Equal(t, 10, n)
	require.Equal(t, data[20:30], readBuf)

	// seek from end
	_, err = lmReader.Seek(0, io.SeekEnd)
	require.NotNil(t, err)
	require.Equal(t, NotSupportSeekEnd, err)

	// not support seeker
	lmReader2 := newWrapLimiterReader(bytes.NewBuffer(data), 100)
	_, err = lmReader2.Seek(0, io.SeekStart)
	require.NotNil(t, err)
	require.Equal(t, NotSupportSeek, err)
}
