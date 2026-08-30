package endpointtest

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"testing"

	"go-api/internal/interfaces/http/handler"
)

const testMaxOpenAPIFileSize = 8 << 20

type fakeOpenAPIFormFile struct {
	size    int64
	openErr error
	content []byte
	readErr error
}

func (f *fakeOpenAPIFormFile) Size() int64 {
	return f.size
}

func (f *fakeOpenAPIFormFile) Open() (multipart.File, error) {
	if f.openErr != nil {
		return nil, f.openErr
	}
	return &fakeMultipartFile{
		Reader:  bytes.NewReader(f.content),
		readErr: f.readErr,
	}, nil
}

type fakeMultipartFile struct {
	*bytes.Reader
	readErr error
	closed  bool
}

func (f *fakeMultipartFile) Close() error {
	f.closed = true
	return nil
}

func (f *fakeMultipartFile) Read(p []byte) (int, error) {
	if f.readErr != nil {
		return 0, f.readErr
	}
	return f.Reader.Read(p)
}

func TestReadOpenAPISpecFromFormFile_Success(t *testing.T) {
	content := []byte(`{"openapi":"3.0.3"}`)
	spec, err := handler.ReadOpenAPISpecFromFormFileForTest(&fakeOpenAPIFormFile{
		size:    int64(len(content)),
		content: content,
	}, testMaxOpenAPIFileSize)
	if err != nil {
		t.Fatalf("read spec: %v", err)
	}
	if !bytes.Equal(spec, content) {
		t.Fatalf("spec: got %q want %q", spec, content)
	}
}

func TestReadOpenAPISpecFromFormFile_HeaderTooLarge(t *testing.T) {
	_, err := handler.ReadOpenAPISpecFromFormFileForTest(&fakeOpenAPIFormFile{
		size:    testMaxOpenAPIFileSize + 1,
		content: []byte("x"),
	}, testMaxOpenAPIFileSize)
	if !errors.Is(err, handler.ErrOpenAPIFileTooLargeForTest()) {
		t.Fatalf("expected errOpenAPIFileTooLarge, got %v", err)
	}
}

func TestReadOpenAPISpecFromFormFile_OpenError(t *testing.T) {
	_, err := handler.ReadOpenAPISpecFromFormFileForTest(&fakeOpenAPIFormFile{
		size:    1,
		openErr: io.ErrClosedPipe,
	}, testMaxOpenAPIFileSize)
	if !errors.Is(err, handler.ErrOpenAPIFileReadForTest()) {
		t.Fatalf("expected errOpenAPIFileRead, got %v", err)
	}
}

func TestReadOpenAPISpecFromFormFile_ReadError(t *testing.T) {
	_, err := handler.ReadOpenAPISpecFromFormFileForTest(&fakeOpenAPIFormFile{
		size:    1,
		content: []byte("x"),
		readErr: io.ErrUnexpectedEOF,
	}, testMaxOpenAPIFileSize)
	if !errors.Is(err, handler.ErrOpenAPIFileReadForTest()) {
		t.Fatalf("expected errOpenAPIFileRead, got %v", err)
	}
}

func TestReadOpenAPISpecFromFormFile_ContentTooLarge(t *testing.T) {
	content := bytes.Repeat([]byte("a"), testMaxOpenAPIFileSize+1)
	_, err := handler.ReadOpenAPISpecFromFormFileForTest(&fakeOpenAPIFormFile{
		size:    testMaxOpenAPIFileSize,
		content: content,
	}, testMaxOpenAPIFileSize)
	if !errors.Is(err, handler.ErrOpenAPIFileTooLargeForTest()) {
		t.Fatalf("expected errOpenAPIFileTooLarge, got %v", err)
	}
}
