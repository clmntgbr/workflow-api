package handler

import (
	"mime/multipart"

	"github.com/gofiber/fiber/v3"
)

// OpenAPIFormFileForTest mirrors openAPIFormFile for endpoint import read tests.
type OpenAPIFormFileForTest interface {
	Open() (multipart.File, error)
	Size() int64
}

// ReadOpenAPISpecFromFormFileForTest exposes readOpenAPISpecFromFormFile.
func ReadOpenAPISpecFromFormFileForTest(file OpenAPIFormFileForTest, maxSize int64) ([]byte, error) {
	return readOpenAPISpecFromFormFile(file, maxSize)
}

// ErrOpenAPIFileTooLargeForTest exposes errOpenAPIFileTooLarge.
func ErrOpenAPIFileTooLargeForTest() error { return errOpenAPIFileTooLarge }

// ErrOpenAPIFileReadForTest exposes errOpenAPIFileRead.
func ErrOpenAPIFileReadForTest() error { return errOpenAPIFileRead }

// ReadOpenAPISpecFromMultipartFileFn is the multipart import read hook type.
type ReadOpenAPISpecFromMultipartFileFn func(*multipart.FileHeader, int64) ([]byte, error)

// SetReadOpenAPISpecFromMultipartFileForTest overrides readOpenAPISpecFromMultipartFile for tests.
func SetReadOpenAPISpecFromMultipartFileForTest(
	fn ReadOpenAPISpecFromMultipartFileFn,
) func() {
	orig := readOpenAPISpecFromMultipartFile
	if fn != nil {
		readOpenAPISpecFromMultipartFile = fn
	}
	return func() {
		readOpenAPISpecFromMultipartFile = orig
	}
}

// RespondQuotaErrorForTest exposes respondQuotaError for isolated mapper tests.
func RespondQuotaErrorForTest(c fiber.Ctx, err error) (bool, error) {
	return respondQuotaError(c, err)
}
