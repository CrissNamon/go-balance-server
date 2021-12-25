package server

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

const (
	BAD_REQUEST_BINDING string = "Wrong request data: %s"

	ERROR_INTERNAL      int = 900
	ERROR_WRONG_REQUEST int = 901
)

type OperationError struct {
	Code int
}

func (opErr *OperationError) Error() string {
	return fmt.Sprintf("Operation completed with code: %d", opErr.Code)
}

func (opErr *OperationError) GetCode() int {
	return opErr.Code
}

func ConvertError(err error) *OperationError {
	switch err.(type) {
	case *OperationError:
		return err.(*OperationError)
	default:
		return &OperationError{ERROR_INTERNAL}
	}
}

func NoRoute(c *gin.Context) {
	r := Result{c, 0, "There is nothing here"}
	r.Response(404)
}
