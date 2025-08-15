package response

import (
	apperrors "marketplace/pkg/errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Responder struct {
	log *logrus.Logger
}

func New(log *logrus.Logger) *Responder {
	return &Responder{log: log}
}

func (r *Responder) Success(c *gin.Context, status int, data interface{}) {
	c.JSON(status, gin.H{
		"success": true,
		"data":    data,
	})
}

func (r *Responder) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func (r *Responder) Error(c *gin.Context, err error) {
	appErr, ok := err.(*apperrors.AppError)
	if !ok {
		r.log.Error("Responder: untyped error: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
		return
	}

	r.log.WithFields(map[string]interface{}{
		"code":    appErr.Code(),
		"message": appErr.Message(),
		"error":   appErr.Error(),
	}).Error("Responder: application error")

	status := mapErrorCodeToStatus(appErr.Code())
	c.JSON(status, gin.H{
		"success": false,
		"error":   appErr.Message(),
	})
}

func mapErrorCodeToStatus(code string) int {
	switch code {
	case "NOT_FOUND":
		return http.StatusNotFound
	case "VALIDATION", "INVALID_TYPE", "INVALID_PAYLOAD":
		return http.StatusBadRequest
	case "INVALID_CREDENTIALS", "INVALID_TOKEN":
		return http.StatusUnauthorized
	case "UPDATE_FAIL", "DELETE_FAIL", "USER_CREATE_FAIL":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
