package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

type ExpectedResultI interface {
	GetStatus(code int) string
	GetHttpCode(code int) int
}

type ExpectedResult struct {
	Statuses  map[int]string
	HttpCodes map[int]int
}

func (er *ExpectedResult) GetStatus(code int) string {
	c, ok := er.Statuses[code]
	if !ok {
		if code == ERROR_INTERNAL {
			return STATUS_INTERNAL_ERROR
		}
		return ""
	}
	return c
}

func (er *ExpectedResult) GetHttpCode(code int) int {
	c, ok := er.HttpCodes[code]
	if !ok {
		if code == ERROR_INTERNAL {
			return 500
		}
		return 200
	}
	return c
}

type Result struct {
	ctx *gin.Context

	Status  int         `json:"status"`
	Message interface{} `json:"data"`
}

func NewResult(c *gin.Context) *Result {
	return &Result{c, STATUS_CODE_OK, ""}
}

func (r *Result) SetStatus(status int) *Result {
	r.Status = status
	return r
}

func (r *Result) SetMessage(message interface{}) *Result {
	r.Message = message
	return r
}

func (r *Result) Give(message interface{}) {
	r.Message = message
	r.Ok()
}

func (r *Result) Ok() {
	r.ctx.JSON(200, r)
}

func (r *Result) Err(err *error, er ExpectedResultI) {
	switch e := (*err).(type) {
	case *OperationError:
		r.SetStatus(e.Code)
		r.SetMessage(er.GetStatus(e.Code))
		r.Response(er.GetHttpCode(e.Code))
	default:
		r.SetStatus(ERROR_INTERNAL)
		r.SetMessage(STATUS_INTERNAL_ERROR)
		r.Response(500)
	}
}

func (r *Result) Response(code int) {
	r.ctx.JSON(code, r)
}

func (r *Result) BadRequest(msg string) {
	r.SetStatus(ERROR_WRONG_REQUEST)
	r.SetMessage(fmt.Sprintf(BAD_REQUEST_BINDING, msg))
	r.Response(400)
}
