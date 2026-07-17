package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// responseBody is the unified JSON shape for all responses:
//
//	{"data": <data>, "code": <code>, "message": <message>}
type responseBody struct {
	Data    interface{} `json:"data"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
}

// Success writes a 200 OK JSON response:
//
//	{"data": <data>, "code": 200, "message": "success"}
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, responseBody{
		Data:    data,
		Code:    http.StatusOK,
		Message: "success",
	})
}

// Error writes a JSON response using the given HTTP status code:
//
//	{"data": null, "code": <code>, "message": <err>}
func Error(c *gin.Context, code int, err string) {
	c.JSON(code, responseBody{
		Data:    nil,
		Code:    code,
		Message: err,
	})
}

// BadRequest writes a 400 Bad Request error response.
func BadRequest(c *gin.Context, err string) {
	Error(c, http.StatusBadRequest, err)
}

// Unauthorized writes a 401 Unauthorized error response.
func Unauthorized(c *gin.Context, err string) {
	Error(c, http.StatusUnauthorized, err)
}

// InternalError writes a 500 Internal Server Error response.
func InternalError(c *gin.Context, err string) {
	Error(c, http.StatusInternalServerError, err)
}
