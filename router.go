package simple_server_runner

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

func (s *ServerRunner) autoBindRouter() error {
	serverTypeOf := reflect.TypeOf(s.server)

	// 获取方法数量
	methodNum := serverTypeOf.NumMethod()
	// 遍历获取所有与公开方法
	for i := 0; i < methodNum; i++ {
		method := serverTypeOf.Method(i)
		// 排除私有方法和Run方法
		_, ok := s.routerWhiteList[method.Name]
		if !method.IsExported() || ok {
			continue
		}

		methodFundType := method.Func.Type()

		//// 检查入参格式 ////
		if methodFundType.NumIn() != 3 {
			return errors.New("the input parameter format is incorrect")
		}
		// 获取第一个参数的类型
		firstInParamType := methodFundType.In(1)

		// 判断第一个参数是否为 *gin.Context 类型
		if firstInParamType != reflect.TypeOf(&gin.Context{}) {
			return errors.New("first input parameter must be *gin.Context")
		}

		// 获取第二个参数的类型
		secondInParamType := methodFundType.In(2)
		// 判断第二个参数是否为结构体类型
		if secondInParamType.Kind() != reflect.Struct {
			return errors.New("second input parameter must be a struct")
		}

		//// 检查出参格式 ////
		if methodFundType.NumOut() != 2 {
			return errors.New("the output parameter format is incorrect")
		}

		// 获取第一个返回值的类型
		firstOutReturnType := methodFundType.Out(0)

		// 判断第一个返回值是否为切片或结构体
		switch firstOutReturnType.Kind() {
		case reflect.Slice, reflect.Struct, reflect.Ptr:
			// 第一个返回值是切片或结构体，继续检查第二个返回值
		default:
			return errors.New("first return value must be a slice or struct")
		}

		// 获取第二个返回值的类型
		secondOutReturnType := methodFundType.Out(1)

		// 判断第二个返回值是否为 error 类型
		if secondOutReturnType != reflect.TypeOf((*error)(nil)).Elem() {
			return errors.New("second return value must be of type error")
		}

		// 创建 Gin 路由处理函数
		handlerFunc := func(c *gin.Context) {
			// 创建参数值的切片
			paramValues := make([]reflect.Value, 3)
			paramValues[0] = reflect.ValueOf(s.server).Elem().Addr()
			paramValues[1] = reflect.ValueOf(c).Elem().Addr()
			paramValues[2] = reflect.New(methodFundType.In(2)).Elem()

			// 绑定请求参数到结构体
			if c.Request.ContentLength > 0 {
				if err := c.ShouldBind(paramValues[2].Addr().Interface()); err != nil {
					c.JSON(http.StatusOK, CommonResponse{
						Code: 500,
						Msg:  err.Error(),
						Data: nil,
					})
					return
				}
			}
			// 绑定uri参数
			if err := c.ShouldBindQuery(paramValues[2].Addr().Interface()); err != nil {
				c.JSON(http.StatusOK, CommonResponse{
					Code: 500,
					Msg:  err.Error(),
					Data: nil,
				})
				return
			}

			// toto 调用函数
			returnValues := method.Func.Call(paramValues)

			// 处理返回值
			var resultValue interface{}
			if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Ptr && returnValues[0].Elem().IsValid() {
				resultValue = returnValues[0].Elem().Interface()
			} else if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Slice {
				resultValue = returnValues[0].Interface()
			}

			errValue, _ := returnValues[1].Interface().(error)
			s.responseFunc(c, resultValue, errValue)
		}
		// 添加路由
		s.gin.Handle("POST", method.Name, handlerFunc)
	}
	return nil
}

func (s *ServerRunner) BindRouter(method, path string, f interface{}) {
	// 检查f是否为一个可调用的函数
	funcType := reflect.TypeOf(f)
	if funcType.Kind() != reflect.Func {
		panic("router bind must be a func")
	}

	// 检查入参是否为 (ctx *gin.Context, request interface{})
	err := s.checkInputParams(funcType)
	if err != nil {
		panic("the function input doesn't match the required format: " + err.Error())
	}

	// 检查出参是否为 (&interface{}, error)
	err = s.checkReturnValues(funcType)
	if err != nil {
		panic("the function return format doesn't match the required format: " + err.Error())
	}

	// 创建 Gin 路由处理函数
	handlerFunc := func(c *gin.Context) {
		// 创建参数值的切片
		paramValues := make([]reflect.Value, 2)
		paramValues[0] = reflect.ValueOf(c).Elem().Addr()
		paramValues[1] = reflect.New(funcType.In(1)).Elem()

		// 绑定请求参数到结构体
		if c.Request.ContentLength > 0 {
			if err := c.ShouldBind(paramValues[1].Addr().Interface()); err != nil {
				c.JSON(http.StatusOK, CommonResponse{
					Code: 500,
					Msg:  err.Error(),
					Data: nil,
				})
				return
			}
		}

		// 绑定uri参数
		if err := c.ShouldBindQuery(paramValues[1].Addr().Interface()); err != nil {
			c.JSON(http.StatusOK, CommonResponse{
				Code: 500,
				Msg:  err.Error(),
				Data: nil,
			})
			return
		}

		// 调用函数
		returnValues := reflect.ValueOf(f).Call(paramValues)

		// 处理返回值
		var resultValue interface{}
		if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Ptr && returnValues[0].Elem().IsValid() {
			resultValue = returnValues[0].Elem().Interface()
		} else if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Slice {
			resultValue = returnValues[0].Interface()
		}
		errValue, _ := returnValues[1].Interface().(error)

		s.responseFunc(c, resultValue, errValue)

	}
	functionName := s.getFunctionName(f)
	s.routerWhiteList[functionName] = struct{}{}
	// 添加路由
	s.gin.Handle(method, path, handlerFunc)
}

// getFunctionName 获取f的方法名
func (s *ServerRunner) getFunctionName(i interface{}) string {
	// 获取函数的指针
	ptr := reflect.ValueOf(i).Pointer()

	// 获取函数的反射信息
	funcName := runtime.FuncForPC(ptr).Name()

	// 使用 strings 包提取方法名
	lastDotIndex := strings.LastIndex(funcName, ".")
	if lastDotIndex != -1 {
		return strings.Replace(funcName[lastDotIndex+1:], "-fm", "", -1)
	}

	return strings.Replace(funcName, "-fm", "", -1)
}

// checkInputParams 检查手动绑定入参
func (s *ServerRunner) checkInputParams(funcType reflect.Type) error {
	// 获取函数的参数个数
	numIn := funcType.NumIn()

	// 判断参数是否符合要求
	if numIn != 2 {
		return errors.New("func must have exactly 2 input parameters")
	}

	// 获取第一个参数的类型
	firstParamType := funcType.In(0)

	// 判断第一个参数是否为 *gin.Context 类型
	if firstParamType != reflect.TypeOf(&gin.Context{}) {
		return errors.New("first input parameter must be *gin.Context")
	}

	// 获取第二个参数的类型
	secondParamType := funcType.In(1)

	// 判断第二个参数是否为结构体类型
	if secondParamType.Kind() != reflect.Struct {
		return errors.New("second input parameter must be a struct")
	}
	return nil
}

// checkReturnValues 检查自动绑定入参
func (s *ServerRunner) checkReturnValues(funcType reflect.Type) error {
	// 获取函数的返回值个数
	numOut := funcType.NumOut()

	// 判断返回值是否符合要求
	if numOut != 2 {
		return errors.New("func must have exactly 2 return values")
	}

	// 获取第一个返回值的类型
	firstReturnType := funcType.Out(0)

	// 判断第一个返回值是否为切片或结构体
	switch firstReturnType.Kind() {
	case reflect.Slice, reflect.Struct, reflect.Ptr:
		// 第一个返回值是切片或结构体，继续检查第二个返回值
	default:
		return errors.New("first return value must be a slice or struct")
	}

	// 获取第二个返回值的类型
	secondReturnType := funcType.Out(1)

	// 判断第二个返回值是否为 error 类型
	if secondReturnType != reflect.TypeOf((*error)(nil)).Elem() {
		return errors.New("second return value must be of type error")
	}

	return nil
}
