package server

import (
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

func (r *Result) give(message interface{}) {
	r.Message = message
	r.ok()
}

func (r *Result) ok() {
	r.ctx.JSON(200, r)
}

func (r *Result) err(err *error) {
	switch e := (*err).(type) {
	case *OperationError:
		r.SetStatus(e.GetCode())
		r.SetMessage(e.Error())
		r.ok()
	default:
		r.SetStatus(STATUS_CODE_INTERNAL_ERROR)
		r.SetMessage(STATUS_INTERNAL_ERROR)
		r.response(500)
	}
}

func (r *Result) response(code int) {
	r.ctx.JSON(code, r)
}
