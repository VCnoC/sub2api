package service

import (
	"bytes"
	"image"
	"image/png"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestTicketFileValidation(t *testing.T) {
	var pngData bytes.Buffer
	require.NoError(t, png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 1, 1))))

	mediaType, err := validateTicketFile(".png", pngData.Bytes())
	require.NoError(t, err)
	require.Equal(t, "image/png", mediaType)

	_, err = validateTicketFile(".json", []byte(`{"ok":`))
	require.ErrorIs(t, err, ErrTicketInvalidFile)

	_, err = validateTicketFile(".txt", []byte{'o', 'k', 0, 'x'})
	require.ErrorIs(t, err, ErrTicketInvalidFile)

	_, err = validateTicketFile(".webp", []byte("RIFF0000WEBP"))
	require.ErrorIs(t, err, ErrTicketInvalidFile)
}

func TestTicketFileStoreSavesPrivateFileAndRejectsLimits(t *testing.T) {
	root := filepath.Join(t.TempDir(), "attachments")
	store := &TicketFileStore{root: root}
	header := ticketTestFileHeader(t, "details.txt", []byte("safe details"))

	items, cleanup, err := store.SaveUploads([]*multipart.FileHeader{header})
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "details.txt", items[0].OriginalName)
	require.NotContains(t, items[0].StorageKey, "details")

	path := filepath.Join(root, items[0].StorageKey)
	stat, err := os.Stat(path)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), stat.Mode().Perm())

	cleanup()
	_, err = os.Stat(path)
	require.ErrorIs(t, err, os.ErrNotExist)

	_, _, err = store.SaveUploads(make([]*multipart.FileHeader, domain.TicketMaxFilesPerReply+1))
	require.ErrorIs(t, err, ErrTicketInvalidFile)

	_, _, err = store.SaveUploads([]*multipart.FileHeader{{Filename: "large.txt", Size: domain.TicketMaxFileBytes + 1}})
	require.ErrorIs(t, err, ErrTicketInvalidFile)

	_, err = store.pathForKey("../outside.txt")
	require.ErrorIs(t, err, ErrTicketInvalidFile)
}

func ticketTestFileHeader(t *testing.T, name string, data []byte) *multipart.FileHeader {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("files", name)
	require.NoError(t, err)
	_, err = part.Write(data)
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	request, err := http.NewRequest(http.MethodPost, "/", &body)
	require.NoError(t, err)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	require.NoError(t, request.ParseMultipartForm(1<<20))
	t.Cleanup(func() { _ = request.MultipartForm.RemoveAll() })
	return request.MultipartForm.File["files"][0]
}
