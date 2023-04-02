package gzip

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

// Compress gzip and base64 encode the source input.
func Compress(source []byte) (string, error) {
	if source == nil {
		return "", nil
	}
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)
	if _, err := gz.Write(source); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

// UnCompress base64 decode and unzip the source input.
func UnCompress(source string) ([]byte, error) {
	if source == "" {
		return nil, nil
	}
	data, err := base64.StdEncoding.DecodeString(source)
	if err != nil {
		return nil, err
	}
	rdata := bytes.NewReader(data)
	r, err := gzip.NewReader(rdata)
	if err != nil {
		return nil, err
	}
	s, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return s, nil
}
