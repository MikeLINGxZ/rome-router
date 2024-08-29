package main

import (
	simple_server_runner "github.com/MikeLINGxZ/rome_router"
	"github.com/gin-gonic/gin"
)

func main() {
	apiRouter := []simple_server_runner.Router{
		{
			Path:        "/api",
			Method:      "POST",
			HandlerFunc: &ApiServer{},
			ChildRouter: []simple_server_runner.Router{
				{
					Path:        "/Login",
					Method:      "POST",
					HandlerFunc: Login,
					ChildRouter: []simple_server_runner.Router{
						{
							Path:   "/Username",
							Method: "GET",
							HandlerFunc: func(ctx *gin.Context) error {
								return nil
							},
							ChildRouter: nil,
							Middlewares: []gin.HandlerFunc{Auth()},
						},
					},
					Middlewares: nil,
				},
			},
			Middlewares: nil,
		}, {},
	}
	router := simple_server_runner.Router{
		Path:        "/",
		Method:      "GET",
		HandlerFunc: Index,
		ChildRouter: apiRouter,
		Middlewares: nil,
	}

	server := simple_server_runner.NewServer()

	server.AddRouter(router)

	err := server.Run(":8787")
	if err != nil {
		panic(err)
	}
}
