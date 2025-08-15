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

func TestWrapLimiterReader_SeekAndRead(t *testing.T) {
	data := make([]byte, 100)
	_, err := rand.Reader.Read(data)
	require.Nil(t, err)
	reader := bytes.NewReader(data)
	lmReader := newWrapLimiterReader(reader, int64(len(data)))

	// 从中间位置开始读取
	startOffset := int64(50)
	_, err = lmReader.Seek(startOffset, io.SeekStart)
	require.Nil(t, err)

	data1, err := io.ReadAll(lmReader)
	require.Nil(t, err)
	require.Equal(t, data[startOffset:], data1)

	// Seek 回到相同位置并重新读取
	_, err = lmReader.Seek(startOffset, io.SeekStart)
	require.Nil(t, err)
	data2, err := io.ReadAll(lmReader)
	require.Nil(t, err)

	// 验证两次读取的数据一致
	require.Equal(t, data1, data2)
}
