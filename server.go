package simple_server_runner

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ServerRunner struct {
	gin             *gin.Engine
	server          interface{}
	routerWhiteList map[string]struct{}
	responseFunc    func(ctx *gin.Context, data interface{}, err error)
}

func NewDefaultServerRunner(server interface{}) *ServerRunner {
	return newServerRunner(gin.Default(), server)
}

func NewServerRunner(engine *gin.Engine, server interface{}) *ServerRunner {
	return newServerRunner(engine, server)
}

func newServerRunner(engine *gin.Engine, server interface{}) *ServerRunner {
	sr := &ServerRunner{gin: engine, server: server}
	sr.initRouterWhileList()
	sr.responseFunc = sr.defaultResponse
	return sr
}

func (s *ServerRunner) Run(addr ...string) error {
	err := s.autoBindRouter()
	if err != nil {
		return err
	}
	return s.gin.Run(addr...)
}

func (s *ServerRunner) Gin() *gin.Engine {
	return s.gin
}

func (s *ServerRunner) CustomResponse(f func(ctx *gin.Context, data interface{}, err error)) {
	s.responseFunc = f
}

// initRouterWhileList 初始化绑定白名单
func (s *ServerRunner) initRouterWhileList() {
	s.routerWhiteList = map[string]struct{}{
		"Run": {},
		"Gin": {},
	}
}

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
func (s *ServerRunner) defaultResponse(ctx *gin.Context, data interface{}, err error) {
	commonResponse := CommonResponse{}
	if err != nil {
		commonResponse.Code = 500
		commonResponse.Msg = err.Error()
	} else {
		commonResponse.Code = 200
		commonResponse.Data = data
	}

	ctx.JSON(http.StatusOK, commonResponse)
}
