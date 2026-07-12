// Package service 实现工单附件的私有本地存储与内容校验。
package service

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	_ "golang.org/x/image/webp"
)

// TicketFileStore 将附件保存在不对公网暴露的 data 目录中。
type TicketFileStore struct{ root string }

func NewTicketFileStore() *TicketFileStore {
	root := strings.TrimSpace(os.Getenv("SUB2API_TICKET_ATTACHMENT_DIR"))
	if root == "" {
		root = filepath.Join("data", "ticket-attachments")
	}
	return &TicketFileStore{root: root}
}

func (s *TicketFileStore) SaveUploads(headers []*multipart.FileHeader) ([]TicketAttachment, func(), error) {
	if len(headers) > domain.TicketMaxFilesPerReply {
		return nil, func() {}, ErrTicketInvalidFile
	}
	if err := os.MkdirAll(s.root, 0o700); err != nil {
		return nil, func() {}, err
	}
	items := make([]TicketAttachment, 0, len(headers))
	created := make([]string, 0, len(headers))
	cleanup := func() {
		for _, key := range created {
			_ = s.Delete(key)
		}
	}
	var total int64
	for _, header := range headers {
		if header == nil || header.Size <= 0 || header.Size > domain.TicketMaxFileBytes {
			cleanup()
			return nil, func() {}, ErrTicketInvalidFile
		}
		total += header.Size
		if total > domain.TicketMaxReplyFileBytes {
			cleanup()
			return nil, func() {}, ErrTicketInvalidFile
		}
		item, err := s.saveOne(header)
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}
		created = append(created, item.StorageKey)
		items = append(items, *item)
	}
	return items, cleanup, nil
}

func (s *TicketFileStore) saveOne(header *multipart.FileHeader) (*TicketAttachment, error) {
	ext := strings.ToLower(filepath.Ext(filepath.Base(header.Filename)))
	if !allowedTicketExtension(ext) {
		return nil, ErrTicketInvalidFile
	}
	src, err := header.Open()
	if err != nil {
		return nil, ErrTicketInvalidFile
	}
	defer src.Close()
	data, err := io.ReadAll(io.LimitReader(src, domain.TicketMaxFileBytes+1))
	if err != nil || len(data) == 0 || len(data) > domain.TicketMaxFileBytes {
		return nil, ErrTicketInvalidFile
	}
	mediaType, err := validateTicketFile(ext, data)
	if err != nil {
		return nil, err
	}
	random := make([]byte, 20)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	key := hex.EncodeToString(random) + ext
	path, err := s.pathForKey(key)
	if err != nil {
		return nil, err
	}
	out, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	if _, err = out.Write(data); err != nil {
		_ = out.Close()
		_ = os.Remove(path)
		return nil, err
	}
	if err = out.Close(); err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	return &TicketAttachment{
		OriginalName: filepath.Base(header.Filename),
		StorageKey:   key,
		MediaType:    mediaType,
		SizeBytes:    int64(len(data)),
	}, nil
}

func (s *TicketFileStore) Open(key string) (*os.File, error) {
	path, err := s.pathForKey(key)
	if err != nil {
		return nil, err
	}
	return os.Open(path)
}

func (s *TicketFileStore) Delete(key string) error {
	path, err := s.pathForKey(key)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *TicketFileStore) pathForKey(key string) (string, error) {
	if key == "" || filepath.Base(key) != key {
		return "", ErrTicketInvalidFile
	}
	root, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}
	path := filepath.Join(root, key)
	if filepath.Dir(path) != root {
		return "", ErrTicketInvalidFile
	}
	return path, nil
}

func allowedTicketExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".txt", ".log", ".json":
		return true
	default:
		return false
	}
}

func validateTicketFile(ext string, data []byte) (string, error) {
	switch ext {
	case ".jpg", ".jpeg":
		if http.DetectContentType(data) != "image/jpeg" {
			return "", ErrTicketInvalidFile
		}
		if _, _, err := image.DecodeConfig(bytes.NewReader(data)); err != nil {
			return "", ErrTicketInvalidFile
		}
		return "image/jpeg", nil
	case ".png":
		if http.DetectContentType(data) != "image/png" {
			return "", ErrTicketInvalidFile
		}
		if _, _, err := image.DecodeConfig(bytes.NewReader(data)); err != nil {
			return "", ErrTicketInvalidFile
		}
		return "image/png", nil
	case ".webp":
		if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WEBP" {
			return "", ErrTicketInvalidFile
		}
		if _, _, err := image.DecodeConfig(bytes.NewReader(data)); err != nil {
			return "", ErrTicketInvalidFile
		}
		return "image/webp", nil
	case ".json":
		if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 || !json.Valid(data) {
			return "", ErrTicketInvalidFile
		}
		return "application/json", nil
	case ".txt", ".log":
		if !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 {
			return "", ErrTicketInvalidFile
		}
		return "text/plain; charset=utf-8", nil
	default:
		return "", ErrTicketInvalidFile
	}
}
