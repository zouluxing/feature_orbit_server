package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	apperr "github.com/zouluxing/ums/pkg/errors"
)

type envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func OK(c *gin.Context, data any)      { c.JSON(http.StatusOK, envelope{Code: 0, Message: "ok", Data: data}) }
func Created(c *gin.Context, data any) { c.JSON(http.StatusCreated, envelope{Code: 0, Message: "created", Data: data}) }
func Fail(c *gin.Context, err error) {
	if ae, ok := err.(*apperr.AppError); ok {
		c.JSON(ae.Status, envelope{Code: ae.Code, Message: ae.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, envelope{Code: 50001, Message: "服务内部错误"})
}
func BadRequest(c *gin.Context, msg string) { c.JSON(http.StatusBadRequest, envelope{Code: 40000, Message: msg}) }
