package util

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "ok", Data: data})
}

type Err struct {
	Status  int
	Message string
}

func (e *Err) Error() string { return e.Message }

func NewErr(status int, message string) *Err {
	return &Err{Status: status, Message: message}
}

func BadRequest(msg string) *Err   { return NewErr(http.StatusBadRequest, msg) }
func Unauthorized(msg string) *Err { return NewErr(http.StatusUnauthorized, msg) }
func Forbidden(msg string) *Err    { return NewErr(http.StatusForbidden, msg) }
func NotFound(msg string) *Err     { return NewErr(http.StatusNotFound, msg) }
func Conflict(msg string) *Err     { return NewErr(http.StatusConflict, msg) }
func BadGateway(msg string) *Err   { return NewErr(http.StatusBadGateway, msg) }
func Internal(msg string) *Err     { return NewErr(http.StatusInternalServerError, msg) }

func Fail(c *gin.Context, err error) {
	var e *Err
	if !errors.As(err, &e) {
		log.Printf("[error] %T: %v", err, err)
		e = Internal("服务器内部错误")
	}
	c.JSON(e.Status, Body{Code: e.Status, Message: e.Message, Data: nil})
}
