package simple_server_runner

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type EmptyRequest struct {
}

type EmptyResponse struct {
}

type CommonResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// defaultResponse 默认返回
func defaultResponse(ctx *gin.Context, data interface{}, errInterface interface{}) {
	commonResponse := CommonResponse{}
	_, ok := errInterface.(*CustomError)
	if ok {
		commonResponse.Code = 500
		commonResponse.Msg = "internal error"

		ctx.JSON(http.StatusOK, commonResponse)
		return
	}
	err := errInterface.(error)
	if err != nil {
		commonResponse.Code = 500
		commonResponse.Msg = err.Error()
	} else {
		commonResponse.Code = 200
		commonResponse.Data = data
	}

	ctx.JSON(http.StatusOK, commonResponse)
}

type CustomError struct {
	s string
}

func (e *CustomError) Error() string {
	return e.s
}

func NewCustomError(text string) *CustomError {
	return &CustomError{text}
}
