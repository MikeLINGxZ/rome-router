package simple_server_runner

import (
	"github.com/gin-gonic/gin"
)

type ServerGroup struct {
	Name        string
	Server      interface{}
	Middlewares []gin.HandlerFunc
}

type ServerRunner struct {
	gin             *gin.Engine
	defaultServer   interface{}
	serverGroup     []ServerGroup
	routerWhiteList map[string]struct{}
	responseFunc    func(ctx *gin.Context, data interface{}, errInterface interface{})
}

func NewDefaultServerRunner(server interface{}) *ServerRunner {
	return &ServerRunner{
		gin:             gin.Default(),
		defaultServer:   server,
		responseFunc:    defaultResponse,
		routerWhiteList: map[string]struct{}{},
	}
}

func NewServerRunner() *ServerRunner {
	return &ServerRunner{
		gin:             gin.Default(),
		responseFunc:    defaultResponse,
		routerWhiteList: map[string]struct{}{},
	}
}

func (s *ServerRunner) Run(addr ...string) error {

	// default
	err := autoBindRouter(s.gin, s.routerWhiteList, ServerGroup{"", s.defaultServer, nil}, s.responseFunc)
	if err != nil {
		return err
	}

	// group
	for _, item := range s.serverGroup {
		item := item
		err := autoBindRouter(s.gin, s.routerWhiteList, item, s.responseFunc)
		if err != nil {
			return err
		}
	}

	return s.gin.Run(addr...)
}

func (s *ServerRunner) Gin() *gin.Engine {
	return s.gin
}

func (s *ServerRunner) AddServerGroup(group ServerGroup) {
	s.serverGroup = append(s.serverGroup, group)
}

func (s *ServerRunner) CustomResponse(f func(ctx *gin.Context, data interface{}, errInterface interface{})) {
	s.responseFunc = f
}
