package upload

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// UpYunService handles UpYun cloud storage uploads.
type UpYunService struct {
	bucket   string
	operator string
	password string
	endpoint string
	domain   string
}

// NewUpYunService creates a new UpYunService with the given config values.
func NewUpYunService(bucket, operator, password, endpoint, domain string) *UpYunService {
	return &UpYunService{
		bucket:   bucket,
		operator: operator,
		password: password,
		endpoint: endpoint,
		domain:   domain,
	}
}

// UploadResult contains the upload result.
type UploadResult struct {
	URL      string `json:"url"`
	Path     string `json:"path"`
	Filename string `json:"filename"`
	MimeType string `json:"mimetype"`
	Size     int64  `json:"size"`
}

// UploadImage uploads an image file to UpYun and returns the result.
func (s *UpYunService) UploadImage(data []byte, filename, mimeType, userFolder string) (*UploadResult, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".jpg"
	}

	// folder = /{openid}/images or /images
	folder := "/images"
	if userFolder != "" {
		folder = "/" + userFolder + "/images"
	}

	filePath := s.generateFilePath(ext, folder)
	result, err := s.upload(data, filePath, mimeType)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		URL:      result.URL,
		Path:     result.Path,
		Filename: filename,
		MimeType: mimeType,
		Size:     int64(len(data)),
	}, nil
}

// UploadFile uploads a generic file to UpYun and returns the result.
func (s *UpYunService) UploadFile(data []byte, filename, mimeType, userFolder string) (*UploadResult, error) {
	ext := strings.ToLower(filepath.Ext(filename))

	// folder = /{openid}/files or /files
	folder := "/files"
	if userFolder != "" {
		folder = "/" + userFolder + "/files"
	}

	filePath := s.generateFilePath(ext, folder)
	result, err := s.upload(data, filePath, mimeType)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		URL:      result.URL,
		Path:     result.Path,
		Filename: filename,
		MimeType: mimeType,
		Size:     int64(len(data)),
	}, nil
}

// generateFilePath generates a unique file path for UpYun storage.
// Format: {prefix}/{timestamp}-{random}.{ext}
func (s *UpYunService) generateFilePath(ext, prefix string) string {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	timestamp := time.Now().UnixMilli()
	random := randomString(6)
	return fmt.Sprintf("%s%d-%s%s", prefix, timestamp, random, ext)
}

// randomString generates a random alphanumeric string of length n.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// uploadResultInternal is used internally by the upload method.
type uploadResultInternal struct {
	URL  string
	Path string
}

// upload performs the actual upload to UpYun.
// Signature algorithm: METHOD&/bucket/path&DATE (no content-length).
func (s *UpYunService) upload(data []byte, filePath, mimeType string) (*uploadResultInternal, error) {
	// Upload URL: http://v0.api.upyun.com/bucket/path
	upURL := fmt.Sprintf("http://%s/%s%s", s.endpoint, s.bucket, filePath)

	date := time.Now().UTC().Format(http.TimeFormat)

	// generateSignature uses filePath (without bucket prefix);
	// internally it prepends /bucket to build the sign URI.
	signature := s.generateSignature("PUT", filePath, date)
	auth := fmt.Sprintf("UPYUN %s:%s", s.operator, signature)

	req, err := http.NewRequest(http.MethodPut, upURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("create upload request: %w", err)
	}

	req.Header.Set("Authorization", auth)
	req.Header.Set("Date", date)
	req.Header.Set("Content-Type", mimeType)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("upload to UpYun: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("upyun upload failed: %s - %s", resp.Status, string(body))
	}

	// Build public URL: domain already includes protocol (e.g. https://upyun.dayangge.site)
	publicURL := fmt.Sprintf("%s%s", s.domain, filePath)
	if s.domain == "" {
		publicURL = upURL
	}

	return &uploadResultInternal{
		URL:  publicURL,
		Path: filePath,
	}, nil
}

// generateSignature generates the UpYun HMAC-SHA1 signature.
// Algorithm: base64(hmac-sha1(md5(password), METHOD&/bucket/path&DATE))
// uri is the file path WITHOUT the bucket prefix; bucket is prepended internally.
func (s *UpYunService) generateSignature(method, uri, date string) string {
	md5Pwd := md5Hash(s.password)
	signURI := "/" + s.bucket + uri
	message := fmt.Sprintf("%s&%s&%s", method, signURI, date)
	mac := hmac.New(sha1.New, []byte(md5Pwd))
	mac.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// md5Hash returns the MD5 hex digest of a string.
func md5Hash(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum(nil))
}
