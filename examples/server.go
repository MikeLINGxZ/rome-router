package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			fmt.Println("x")
			c.JSON(http.StatusUnauthorized, gin.H{})
			c.Abort()
		}
	}
}

func Index(c *gin.Context) error {
	return nil
}

func IndexError(c *gin.Context) error {
	return errors.New("error")
}

func IndexParamsError(c *gin.Context, req IndexRequest) (*IndexResponse, error) {
	fmt.Println("Say:", req.Say)
	return &IndexResponse{
		Back: "back a msg",
	}, nil
}

type ApiServer struct {
}

func Login(ctx *gin.Context, req LoginRequest) (*LoginResponse, error) {
	fmt.Println("Username:", req.Username, ", Password:", req.Password)
	return &LoginResponse{Token: "EBF09F808D8C4049B8F03B4FAAE1D223"}, nil
}

func (a *ApiServer) LoginOut(ctx *gin.Context) error {
	return errors.New("this is a error")
}
