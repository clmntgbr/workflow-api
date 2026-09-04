package handler

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
)

type openAPIFormFile interface {
	Open() (multipart.File, error)
	Size() int64
}

type multipartFormFile struct {
	header *multipart.FileHeader
}

func (f multipartFormFile) Open() (multipart.File, error) {
	return f.header.Open()
}

func (f multipartFormFile) Size() int64 {
	return f.header.Size
}

var (
	errOpenAPIFileTooLarge = errors.New("openapi file too large")
	errOpenAPIFileRead     = errors.New("openapi file read failed")

	readOpenAPISpecFromMultipartFile = defaultReadOpenAPISpecFromMultipartFile
)

func defaultReadOpenAPISpecFromMultipartFile(fileHeader *multipart.FileHeader, maxSize int64) ([]byte, error) {
	return readOpenAPISpecFromFormFile(multipartFormFile{header: fileHeader}, maxSize)
}

func readOpenAPISpecFromFormFile(file openAPIFormFile, maxSize int64) ([]byte, error) {
	if file.Size() > maxSize {
		return nil, errOpenAPIFileTooLarge
	}

	f, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errOpenAPIFileRead, err)
	}
	defer f.Close()

	spec, err := io.ReadAll(io.LimitReader(f, maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errOpenAPIFileRead, err)
	}
	if int64(len(spec)) > maxSize {
		return nil, errOpenAPIFileTooLarge
	}

	return spec, nil
}
