package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	BAD_REQUEST_BINDING string = "Wrong request data: %s"
)

type OperationError struct {
	Code int
}

func (opErr *OperationError) Error() string {
	if msg, ok := STATUS[opErr.Code]; !ok {
		return fmt.Sprintf("Operation completed with code: %d", opErr.Code)
	} else {
		return msg
	}
}

func (opErr *OperationError) GetCode() int {
	return opErr.Code
}

func NoRoute(c *gin.Context) {
	r := Result{c, 0, "There is nothing here"}
	r.response(404)
}
