package simple_server_runner

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/url"
	"reflect"
	"runtime"
	"strings"
)

func autoBindRouter(ginEngine *gin.Engine, whileRouter map[string]struct{}, serverGroup ServerGroup, responseFunc func(ctx *gin.Context, data, err interface{})) error {
	if serverGroup.Server == nil {
		return nil
	}
	serverTypeOf := reflect.TypeOf(serverGroup.Server)

	// 获取方法数量
	methodNum := serverTypeOf.NumMethod()

	ginGroup := ginEngine.Group(serverGroup.Name).Use(serverGroup.Middlewares...)

	// 遍历获取所有与公开方法
	for i := 0; i < methodNum; i++ {
		// 获取方法
		method := serverTypeOf.Method(i)

		// 排除私有方法
		if !method.IsExported() {
			continue
		}

		routerUri, err := url.JoinPath("/", serverGroup.Name, method.Name)
		if err != nil {
			return err
		}
		fmt.Println(routerUri)
		// 排除已注册路由
		_, ok := whileRouter[routerUri]
		if ok {
			continue
		}

		// 获取函数信息
		methodFundType := method.Func.Type()

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
			paramValues[0] = reflect.ValueOf(serverGroup.Server).Elem().Addr()
			paramValues[1] = reflect.ValueOf(c).Elem().Addr()
			if methodFundType.NumIn() == 3 {
				paramValues[2] = reflect.New(methodFundType.In(2)).Elem()
				// 绑定请求参数到结构体
				if c.Request.ContentLength > 0 {
					if err := c.ShouldBind(paramValues[2].Addr().Interface()); err != nil {
						responseFunc(c, nil, err)
						return
					}
				}
				// 绑定uri参数
				if err := c.ShouldBindQuery(paramValues[2].Addr().Interface()); err != nil {
					responseFunc(c, nil, err)
					return
				}
			}

			// 处理返回值
			var resultValue interface{}
			var errInterface interface{}

			// 调用函数
			returnValues := method.Func.Call(paramValues)

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

			responseFunc(c, resultValue, errInterface)
		}
		ginGroup.Handle("POST", method.Name, handlerFunc)
		// 添加路由
		whileRouter[routerUri] = struct{}{}
	}
	return nil
}

func (s *ServerRunner) BindRouter(method, path string, f interface{}, middlewares []gin.HandlerFunc) {
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
		paramValues := make([]reflect.Value, funcType.NumIn())
		paramValues[0] = reflect.ValueOf(c).Elem().Addr()
		if funcType.NumIn() == 2 {
			paramValues[1] = reflect.New(funcType.In(1)).Elem()
			// 绑定请求参数到结构体
			if c.Request.ContentLength > 0 {
				if err := c.ShouldBind(paramValues[1].Addr().Interface()); err != nil {
					s.responseFunc(c, nil, err)
					return
				}
			}
			// 绑定uri参数
			if err := c.ShouldBindQuery(paramValues[1].Addr().Interface()); err != nil {
				s.responseFunc(c, nil, err)
				return
			}
		}

		// 调用函数
		returnValues := reflect.ValueOf(f).Call(paramValues)

		// 处理返回值
		var resultValue interface{}
		var errInterface interface{}

		if funcType.NumOut() == 1 {
			errInterface = returnValues[0].Interface()
		} else if funcType.NumOut() == 2 {
			if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Ptr && returnValues[0].Elem().IsValid() {
				resultValue = returnValues[0].Elem().Interface()
			} else if returnValues[0].IsValid() && returnValues[0].Kind() == reflect.Slice {
				resultValue = returnValues[0].Interface()
			}
			errInterface = returnValues[1].Interface()
		}

		s.responseFunc(c, resultValue, errInterface)

	}
	// 添加路由
	s.gin.Handle(method, path, append([]gin.HandlerFunc{handlerFunc}, middlewares...)...)
	s.routerWhiteList[path] = struct{}{}
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
func (s *ServerRunner) checkReturnValues(funcType reflect.Type) error {
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
	} else if numOut == 2 {

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
