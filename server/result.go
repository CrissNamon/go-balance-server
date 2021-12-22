package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

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

func (r *Result) Err(err *error) {
	switch e := (*err).(type) {
	case *OperationError:
		r.SetStatus(e.GetCode())
		r.SetMessage(e.Error())
		r.Ok()
	default:
		r.SetStatus(STATUS_CODE_INTERNAL_ERROR)
		r.SetMessage(STATUS_INTERNAL_ERROR)
		r.Response(500)
	}
}

func (r *Result) Response(code int) {
	r.ctx.JSON(code, r)
}

func (r *Result) BadRequest(msg string) {
	r.SetStatus(STATUS_CODE_WRONG_REQUEST)
	r.SetMessage(fmt.Sprintf(BAD_REQUEST_BINDING, msg))
	r.Response(400)
}
