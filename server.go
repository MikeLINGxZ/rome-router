package simple_server_runner

import (
	"errors"
	"github.com/gin-gonic/gin"
	"reflect"
)

type ResponseHandler func(c *gin.Context, response interface{}, err interface{})

func defaultResponseHandler(c *gin.Context, response interface{}, err interface{}) {
	responseMap := gin.H{}
	responseMap["error"] = err
	responseMap["data"] = response
	c.JSON(200, responseMap)
}

// Router a router config
type Router struct {
	Path        string            // url path
	Method      string            // method of this handler
	HandlerFunc interface{}       // handler func or func group
	ChildRouter []Router          // child router
	Middlewares []gin.HandlerFunc // middleware for this handler
}

// Server is the main server
type Server struct {
	routers  []Router        // routers
	response ResponseHandler // response func
	engine   *gin.Engine     // gin
}

// NewServer init an empty server
func NewServer() *Server {
	return &Server{engine: gin.Default(), response: defaultResponseHandler}
}

// NewServerWithRouters init a server with routers
func NewServerWithRouters(routers []Router) *Server {
	return &Server{engine: gin.Default(), routers: routers, response: defaultResponseHandler}
}

// AddRouter add a router to server
func (s *Server) AddRouter(router Router) {
	s.routers = append(s.routers, router)
}

// AddRouters add routers to server
func (s *Server) AddRouters(routers ...Router) {
	s.routers = append(s.routers, routers...)
}

// SetEngine set a gin engine
func (s *Server) SetEngine(engine *gin.Engine) {
	s.engine = engine
}

// SetResponse set a response handler
func (s *Server) SetResponse(response ResponseHandler) {
	s.response = response
}

// Run server
func (s *Server) Run(addr ...string) error {
	err := s.initRouter()
	if err != nil {
		return err
	}
	return s.engine.Run(addr...)
}

func (s *Server) initRouter() error {

	var init func(engine *gin.RouterGroup, parentRouter Router, middlewares ...gin.HandlerFunc) error

	init = func(groupEngine *gin.RouterGroup, parentRouter Router, middlewares ...gin.HandlerFunc) error {
		groupEngine.Use(middlewares...)
		// register handler server
		if parentRouter.HandlerFunc != nil {
			// get handlerFunc type
			handlerFuncType := reflect.TypeOf(parentRouter.HandlerFunc)
			switch handlerFuncType.Kind() {
			case reflect.Func:
				err := s.bindRouter(groupEngine, "", parentRouter.Method, parentRouter.HandlerFunc)
				if err != nil {
					return err
				}
			case reflect.Ptr, reflect.Struct:
				err := s.bindGroupRouter(groupEngine, parentRouter.Path, parentRouter.Method, parentRouter.HandlerFunc)
				if err != nil {
					return err
				}
			default:
				return errors.New("unsupported handler func")
			}
		}

		// register child router
		for _, childRouter := range parentRouter.ChildRouter {
			newGroupEngine := groupEngine.Group(childRouter.Path)
			err := init(newGroupEngine, childRouter, childRouter.Middlewares...)
			if err != nil {
				return err
			}
		}

		return nil
	}

	for _, item := range s.routers {
		group := s.engine.Group(item.Path)
		err := init(group, item, item.Middlewares...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) bindRouter(engine *gin.RouterGroup, path, method string, f interface{}, middlewares ...gin.HandlerFunc) error {

	// get type of f
	handlerFuncType := reflect.TypeOf(f)

	// check handler func in type
	err := checkInputParams(handlerFuncType)
	if err != nil {
		return err
	}

	// check handler func return type
	err = checkReturnValues(handlerFuncType)
	if err != nil {
		return err
	}

	handlerFunc := func(c *gin.Context) {
		// caller param
		paramValues := make([]reflect.Value, handlerFuncType.NumIn())
		paramValues[0] = reflect.ValueOf(c).Elem().Addr()
		if handlerFuncType.NumIn() == 2 {
			paramValues[1] = reflect.New(handlerFuncType.In(1)).Elem()
			// 绑定请求参数到结构体
			if c.Request.ContentLength > 0 {
				if err := c.ShouldBind(paramValues[1].Addr().Interface()); err != nil {
					s.response(c, nil, err)
					return
				}
			}
			// 绑定uri参数
			if err := c.ShouldBindQuery(paramValues[1].Addr().Interface()); err != nil {
				s.response(c, nil, err)
				return
			}
		}

		// 调用函数
		returnValues := reflect.ValueOf(f).Call(paramValues)

		// 处理返回值
		var resultValue interface{}
		var errInterface interface{}

		if handlerFuncType.NumOut() == 1 {
			errInterface = returnValues[0].Interface()
		} else if handlerFuncType.NumOut() == 2 {
			if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Ptr && returnValues[0].Elem().IsValid() {
				resultValue = returnValues[0].Elem().Interface()
			} else if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Slice {
				resultValue = returnValues[0].Interface()
			}
			errInterface = returnValues[1].Interface()
		}

		s.response(c, resultValue, errInterface)

	}

	// 添加路由
	engine.Handle(method, path, append([]gin.HandlerFunc{handlerFunc}, middlewares...)...)

	return nil
}

func (s *Server) bindGroupRouter(engine *gin.RouterGroup, path, method string, f interface{}, middlewares ...gin.HandlerFunc) error {
	group := reflect.TypeOf(f)
	methodNum := group.NumMethod()
	for i := 0; i < methodNum; i++ {
		// 获取方法
		methodFunc := group.Method(i)
		// 排除私有方法
		if !methodFunc.IsExported() {
			continue
		}
		// 获取函数信息
		methodFundType := methodFunc.Func.Type()

		// 检查入参数是否符合
		if methodFundType.NumIn() != 2 && methodFundType.NumIn() != 3 {
			continue
		}

		// 判断第一个参数是否为 *gin.Context 类型
		if methodFundType.In(1) != reflect.TypeOf(&gin.Context{}) {
			continue
		}

		// 如果参数有3个，代表有入参结构体
		if methodFundType.NumIn() == 3 && methodFundType.In(2).Kind() != reflect.Struct {
			continue
		}

		//// 检查出参格式 ////
		if methodFundType.NumOut() != 1 && methodFundType.NumOut() != 2 {
			continue
		}

		if methodFundType.NumOut() == 1 {
			if methodFundType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
				return errors.New("sign return value must be of type error")
			}
		} else if methodFundType.NumOut() == 2 {
			// 判断第一个返回值是否为切片或结构体
			switch methodFundType.Out(0).Kind() {
			case reflect.Slice, reflect.Struct, reflect.Ptr:
				// 第一个返回值是切片或结构体，继续检查第二个返回值
			default:
				return errors.New("first return value must be a slice or struct")
			}
			// 判断第二个返回值是否为 error 类型
			if methodFundType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
				return errors.New("second return value must be of type error")
			}
		}

		// 创建 Gin 路由处理函数
		handlerFunc := func(c *gin.Context) {
			// 创建参数值的切片
			paramValues := make([]reflect.Value, 3)
			paramValues[0] = reflect.ValueOf(f).Elem().Addr()
			paramValues[1] = reflect.ValueOf(c).Elem().Addr()
			if methodFundType.NumIn() == 3 {
				paramValues[2] = reflect.New(methodFundType.In(2)).Elem()
				// 绑定请求参数到结构体
				if c.Request.ContentLength > 0 {
					if err := c.ShouldBind(paramValues[2].Addr().Interface()); err != nil {
						s.response(c, nil, err)
						return
					}
				}
				// 绑定uri参数
				if err := c.ShouldBindQuery(paramValues[2].Addr().Interface()); err != nil {
					s.response(c, nil, err)
					return
				}
			}

			// 处理返回值
			var resultValue interface{}
			var errInterface interface{}

			// 调用函数
			returnValues := methodFunc.Func.Call(paramValues)

			// 判断返回
			if methodFundType.NumOut() == 1 {
				errInterface = returnValues[0].Interface()
			} else if methodFundType.NumOut() == 2 {
				if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Ptr && returnValues[0].Elem().IsValid() {
					resultValue = returnValues[0].Elem().Interface()
				} else if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Slice {
					resultValue = returnValues[0].Interface()
				}
				errInterface = returnValues[1].Interface()
			}

			s.response(c, resultValue, errInterface)
		}
		engine.Handle(method, methodFunc.Name, handlerFunc)
	}

	return nil
}

// checkInputParams check handlerFunc params
func checkInputParams(funcType reflect.Type) error {
	// 获取函数的参数个数
	numIn := funcType.NumIn()

	// 判断参数是否符合要求
	if numIn != 1 && numIn != 2 {
		return errors.New("func must have exactly 2 input parameters")
	}

	// 获取第一个参数的类型
	firstParamType := funcType.In(0)

	// 判断第一个参数是否为 *gin.Context 类型
	if firstParamType != reflect.TypeOf(&gin.Context{}) {
		return errors.New("first input parameter must be *gin.Context")
	}

	if numIn == 2 && funcType.In(1).Kind() != reflect.Struct {
		return errors.New("second input parameter must be a struct")
	}

	return nil
}

// checkReturnValues 检查自动绑定入参
func checkReturnValues(funcType reflect.Type) error {
	// 获取函数的返回值个数
	numOut := funcType.NumOut()

	// 判断返回值是否符合要求
	if numOut != 1 && numOut != 2 {
		return errors.New("func must have exactly 2 return values")
	}

	if numOut == 1 {
		if funcType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			return errors.New("second return value must be of type error")
		}
	} else {

		// 判断第一个返回值是否为切片或结构体
		switch funcType.Out(0).Kind() {
		case reflect.Slice, reflect.Struct, reflect.Ptr:
			// 第一个返回值是切片或结构体，继续检查第二个返回值
		default:
			return errors.New("first return value must be a slice or struct")
		}

		// 判断第二个返回值是否为 error 类型
		if funcType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
			return errors.New("second return value must be of type error")
		}
	}

	return nil
}
