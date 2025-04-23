package tos

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"hash"
	"hash/crc64"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	trailerHeaderSeparator               = ":"
	crlf                                 = "\r\n"
	contentEncodingHeaderName            = "content-encoding"
	tosChunkedContentEncodingHeaderValue = "tos-chunked"
	tosTrailerHeaderName                 = "x-tos-trailer"
	tosChecksumCrc64Header               = "x-tos-hash-crc64ecma"
	defaultChunkSize                     = 64 * 1024
)

var (
	crlfBytes       = []byte(crlf)
	lastChunkBytes  = []byte("0" + crlf)
	lastChunkLength = int64(len(lastChunkBytes))
	crlBytesLength  = int64(len(crlfBytes))
)

type tosChunkEncodingReader struct {
	body          io.Reader
	trailerLength int64
	contentLength int64
	trailerHeader map[string]trailerValue
}

func newTosChunkReader(contentLength int64, body io.Reader) io.Reader {
	var endChunk bytes.Buffer
	if contentLength == 0 {
		endChunk.Write(lastChunkBytes)
		return &endChunk
	}
	endChunk.WriteString(crlf)
	endChunk.Write(lastChunkBytes)
	var header bytes.Buffer
	header.WriteString(strconv.FormatInt(contentLength, 16))
	header.WriteString(crlf)
	return io.MultiReader(
		&header,
		body,
		&endChunk,
	)
}

type trailerValue interface {
	GetValue() (string, error)
	GetLength() int64
}

func newTosChunkEncodingReader(contentLength int64, trailerHeader map[string]trailerValue, body io.Reader) *tosChunkEncodingReader {
	var reader io.Reader
	if contentLength == -1 {
		reader = newBufferedTOSChunkReader(body)

	} else {
		reader = newTosChunkReader(contentLength, body)
	}
	tr := newTrailerReader(trailerHeader)
	return &tosChunkEncodingReader{body: io.MultiReader(reader, tr, bytes.NewBuffer(crlfBytes)), trailerLength: tr.GetLength(), contentLength: contentLength, trailerHeader: trailerHeader}
}

func (t *tosChunkEncodingReader) getLength() int64 {
	if t.contentLength == -1 {
		return -1
	}
	bodyLength := int64(0)
	if t.contentLength != 0 {
		contentLengthStr := strconv.FormatInt(t.contentLength, 16)
		bodyLength += int64(len(contentLengthStr)) + crlBytesLength + crlBytesLength + t.contentLength

	}
	return bodyLength + t.trailerLength + lastChunkLength + crlBytesLength
}

func (t *tosChunkEncodingReader) getHttpHeader() http.Header {
	httpHeader := make(http.Header)
	httpHeader.Add(contentEncodingHeaderName, tosChunkedContentEncodingHeaderValue)
	httpHeader.Add("x-tos-content-sha256", "STREAMING-UNSIGNED-PAYLOAD-TRAILER")
	for k := range t.trailerHeader {
		httpHeader.Add(tosTrailerHeaderName, k)
	}
	return httpHeader
}

type trailerReader struct {
	s             *bytes.Buffer
	trailerHeader map[string]trailerValue
	length        int64
}

func newTrailerReader(trailerHeader map[string]trailerValue) *trailerReader {
	length := int64(0)
	for key, value := range trailerHeader {
		length += int64(len(key)+len(trailerHeaderSeparator)) + value.GetLength() + int64(len(crlf))
	}
	return &trailerReader{trailerHeader: trailerHeader, length: length}
}

func (t *trailerReader) GetLength() int64 {
	return t.length
}

func (t *trailerReader) Read(p []byte) (n int, err error) {
	if len(t.trailerHeader) == 0 {
		return 0, io.EOF
	}
	if t.s == nil {
		buff := bytes.NewBuffer(nil)
		for key, trailer := range t.trailerHeader {
			buff.WriteString(key)
			buff.WriteString(trailerHeaderSeparator)
			v, err := trailer.GetValue()
			if err != nil {
				return 0, errors.New("Get trailer value failed, err: " + err.Error())
			}
			buff.WriteString(v)
			buff.WriteString(crlf)
		}
		t.s = buff
	}
	return t.s.Read(p)
}

type tosBufferChunkReader struct {
	reader       io.Reader
	chunkSize    int64
	chunkSizeStr string

	headerBuffer   *bytes.Buffer
	bodyBuffer     *bytes.Buffer
	multiReader    io.Reader
	multiReaderLen int
	endChunkDone   bool
}

func newBufferedTOSChunkReader(reader io.Reader) *tosBufferChunkReader {
	chunkSize := int64(defaultChunkSize)
	return &tosBufferChunkReader{
		reader:       reader,
		chunkSize:    chunkSize,
		chunkSizeStr: strconv.FormatInt(int64(defaultChunkSize), 16),

		headerBuffer: bytes.NewBuffer(make([]byte, 0, 64)),
		bodyBuffer:   bytes.NewBuffer(make([]byte, 0, int(chunkSize)+len(crlf))),
	}
}

func (t *tosBufferChunkReader) Read(p []byte) (n int, err error) {
	if t.multiReaderLen == 0 && t.endChunkDone {
		return 0, io.EOF
	}
	if t.multiReader == nil || t.multiReaderLen == 0 {
		t.multiReader, t.multiReaderLen, err = t.newMultiReader()
		if err != nil {
			return 0, err
		}
	}

	n, err = t.multiReader.Read(p)
	t.multiReaderLen -= n

	if err == io.EOF && !t.endChunkDone {
		err = nil
	}
	return n, err
}

func (t *tosBufferChunkReader) newMultiReader() (io.Reader, int, error) {
	n, err := io.Copy(t.bodyBuffer, io.LimitReader(t.reader, t.chunkSize))
	if err != nil {
		return nil, 0, err
	}
	if n == 0 {
		t.headerBuffer.Reset()
		t.headerBuffer.WriteString("0")
		t.headerBuffer.WriteString(crlf)
		t.endChunkDone = true
		return t.headerBuffer, t.headerBuffer.Len(), nil
	}
	t.bodyBuffer.WriteString(crlf)

	chunkSizeStr := t.chunkSizeStr
	if n != t.chunkSize {
		chunkSizeStr = strconv.FormatInt(n, 16)
	}

	t.headerBuffer.Reset()
	t.headerBuffer.WriteString(chunkSizeStr)
	t.headerBuffer.WriteString(crlf)

	return io.MultiReader(
		t.headerBuffer,
		t.bodyBuffer,
	), t.headerBuffer.Len() + t.bodyBuffer.Len(), nil
}

type chunkReader struct {
	br            *bufio.Reader
	cur           io.Reader
	src           io.ReadCloser
	crlf          []byte
	crcChecker    hash.Hash64
	trailerHeader http.Header
}

func (c *chunkReader) Close() error {
	return c.src.Close()
}

func newChunkReader(input io.ReadCloser, realContentLength int64) io.ReadCloser {
	cr := &chunkReader{
		src:           input,
		br:            bufio.NewReader(input),
		crcChecker:    crc64.New(crc64.MakeTable(crc64.ECMA)),
		trailerHeader: make(http.Header),
		crlf:          make([]byte, 2),
	}
	cr.cur = io.TeeReader(io.LimitReader(cr.br, int64(realContentLength)), cr.crcChecker)

	return cr

}

func (c *chunkReader) readCRLF() (err error) {
	_, err = io.ReadFull(c.br, c.crlf[:2])
	if err != nil {
		return err
	}
	if c.crlf[0] != '\r' || c.crlf[1] != '\n' {
		return newTosClientError("malformed chunked encoding", err)
	}
	c.crlf = c.crlf[:0]
	return nil
}

func (c *chunkReader) ReadTrailersHeader() error {
	var prefix []byte
	for {
		line, isPrefix, err := c.br.ReadLine()
		if err != nil {
			return err
		}
		if len(line) == 0 {
			return nil
		}
		prefix = append(prefix, line...)
		if !isPrefix {
			lineStr := strings.TrimSpace(string(prefix))
			lineSplit := strings.Split(lineStr, ":")
			if len(lineSplit) != 2 {
				return newTosClientError("malformed chunked encoding", err)
			}
			c.trailerHeader.Set(lineSplit[0], lineSplit[1])
			prefix = prefix[:0]
		}
	}
}

func (c *chunkReader) Read(p []byte) (n int, err error) {
	n, err = c.cur.Read(p)
	if err == io.EOF {
		err = c.readCRLF()
		if err != nil {
			return 0, err
		}
		err = c.ReadTrailersHeader()
		if err != nil {
			return 0, err
		}
		if base64.StdEncoding.EncodeToString(c.crcChecker.Sum(nil)) != c.trailerHeader.Get(tosChecksumCrc64Header) {
			return 0, newTosClientError("Crc64 not equal", nil)
		}
		return n, io.EOF
	} else if err != nil {
		return n, err
	}
	return n, err
}
