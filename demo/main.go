package main

import (
	"errors"
	"fmt"
	simple_server_runner "github.com/MikeLINGxZ/simple-server-runner"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type CustomResponse struct {
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data"`
}

func main() {
	runner := simple_server_runner.NewDefaultServerRunner(&Server{})

	runner.BindRouter("GET", "GetAge", GetAge)
	runner.BindRouter("GET", "NothingToDoWithAuto", NothingToDoWithAuto)
	runner.BindRouter("GET", "ErrorWithAuto", ErrorWithAuto)
	runner.BindRouter("GET", "CustomErrorWithAuto", CustomErrorWithAuto)

	runner.CustomResponse(func(ctx *gin.Context, data interface{}, errInterface interface{}) {
		resp := CustomResponse{}
		if errInterface != nil {
			err, ok := errInterface.(*CustomResponse)
			if ok {
				resp.Error = err.Error
			} else {
				resp.Error = "gg"
			}
			ctx.JSON(http.StatusBadGateway, resp)
			return
		}
		resp.Data = data
		ctx.JSON(http.StatusOK, resp)
		return
	})

	err := runner.Run(":9003")
	if err != nil {
		panic(err)
	}
}

type Server struct {
}

type GetUserRequest struct {
	UserName string `json:"user_name"`
}

type GetUserResponse struct {
	Msg string `json:"msg"`
}

func (s *Server) GetUser(ctx *gin.Context, req GetUserRequest) (*GetUserResponse, error) {
	return &GetUserResponse{Msg: "hello,im " + req.UserName}, nil
}

func (s *Server) NothingToDo(ctx *gin.Context) error {
	log.Println("NothingToDo")
	return nil
}

type GetAgeRequest struct {
	UserName *string `form:"user_name"`
}

type GetAgeResponse struct {
	Msg string `json:"msg"`
}

func GetAge(ctx *gin.Context, req GetAgeRequest) (*GetAgeResponse, error) {
	if req.UserName == nil || *req.UserName == "" {
		return nil, errors.New("user name is nil")
	}
	return &GetAgeResponse{Msg: fmt.Sprintf("%s is 20 years old", req.UserName)}, nil
}

func NothingToDoWithAuto(ctx *gin.Context) error {
	log.Println("NothingToDo")
	return nil
}

func CustomErrorWithAuto(ctx *gin.Context) error {
	log.Println("NothingToDo")
	return simple_server_runner.NewCustomError("test error")
}

func ErrorWithAuto(ctx *gin.Context) error {
	log.Println("NothingToDo")
	return errors.New("test error")
}
