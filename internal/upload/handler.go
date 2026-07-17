package upload

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"go-service/pkg/response"
)

// uploadServiceIface abstracts upload service for handler injection.
type uploadServiceIface interface {
	UploadImage(data []byte, filename, mimeType, userFolder string) (*UploadResult, error)
	UploadFile(data []byte, filename, mimeType, userFolder string) (*UploadResult, error)
}

// UploadHandler handles HTTP requests for file uploads.
type UploadHandler struct {
	service uploadServiceIface
}

// NewUploadHandler creates a new UploadHandler.
func NewUploadHandler(service *UpYunService) *UploadHandler {
	return &UploadHandler{service: service}
}

// allowedImageTypes maps MIME types to their extensions.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

const (
	maxImageSize = 10 * 1024 * 1024 // 10MB
	maxFileSize  = 50 * 1024 * 1024 // 50MB
)

// getUserFolder extracts a user-specific folder path from the context.
// Uses the userID if available.
func getUserFolder(c *gin.Context) string {
	val, ok := c.Get("userID")
	if !ok {
		return ""
	}
	uid, ok := val.(uint)
	if !ok || uid == 0 {
		return ""
	}
	return fmt.Sprintf("%d", uid)
}

// UploadImage godoc
// @Summary Upload an image
// @Tags upload
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 413 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /upload/image [post]
func (h *UploadHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	if header.Size > maxImageSize {
		response.Error(c, http.StatusRequestEntityTooLarge, "file too large")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if !allowedImageTypes[contentType] {
		response.BadRequest(c, "invalid image type")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		response.InternalError(c, "failed to read file")
		return
	}

	userFolder := getUserFolder(c)
	result, err := h.service.UploadImage(data, filepath.Base(header.Filename), contentType, userFolder)
	if err != nil {
		fmt.Printf("[Upload Error] UploadImage failed: %v\n", err)
		response.InternalError(c, "upload failed: "+err.Error())
		return
	}

	response.Success(c, result)
}

// UploadFile godoc
// @Summary Upload a file
// @Tags upload
// @Security Bearer
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 413 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /upload/file [post]
func (h *UploadHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "file is required")
		return
	}

	if header.Size > maxFileSize {
		response.Error(c, http.StatusRequestEntityTooLarge, "file too large")
		return
	}

	contentType := header.Header.Get("Content-Type")

	data, err := io.ReadAll(file)
	if err != nil {
		response.InternalError(c, "failed to read file")
		return
	}

	userFolder := getUserFolder(c)
	result, err := h.service.UploadFile(data, filepath.Base(header.Filename), contentType, userFolder)
	if err != nil {
		fmt.Printf("[Upload Error] UploadFile failed: %v\n", err)
		response.InternalError(c, "upload failed: "+err.Error())
		return
	}

	response.Success(c, result)
}
