# simple-server-runner
a simple framework base gin that you can esay to code your web server

## What is that?
when I use gin to codding my web server,i always need to bind request and write response.I think,if we can codding web api like codding proto server,that will be easy and distinct.

In order to make codding web api easy, I write this repo.when I use it,I just need to define a struct and some method,the framework will auto bind to router use POST with JSON.

method format like:  func (ctx *gin.Context,req {struct}) (resp {struct ptr},err error)  
**OR**  
autoBind will router the method that param start with `*gin.Context` and params less than 2

## How to use?

### install
```shell
go get github.com/MikeLINGxZ/simple-server-runner
```
### define your server
```go
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
```
### init your default server and run
```go
runner := simple_server_runner.NewDefaultServerRunner(&Server{})
runner.Run()
```
### OR
### init your custom server and run
```go
runner := simple_server_runner.NewServerRunner()
runner.AddServerGroup(ssr.ServerGroup{
    Name:        "api",
    Server:      &Server{},
    Middlewares: []gin.HandlerFunc{Auth, Cors},
}) // url: /api/{method name}
runner.BindRouter("POST", "/api/Login", (&Server{}).Login, []gin.HandlerFunc{Cors})
runner.Run()
```
you will see
```
[GIN-debug] POST   /api/GetUser              --> github.com/MikeLINGxZ/simple-server-runner.autoBindRouter.func1 (3 handlers)
[GIN-debug] POST   /api/NothingToDo          --> github.com/MikeLINGxZ/simple-server-runner.autoBindRouter.func1 (3 handlers)
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
[GIN-debug] Listening and serving HTTP on :9003
```

## Others

### ManualBindingRouter
#### define your func
```go
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

```
#### bind 
```go
runner.BindRouter("GET", "GetAge", GetAge)
```

### Custom your response
```go
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
```

### Empty request OR Empty response
you don't have to define an empty struct yourself.
```go
func GetAge(ctx *gin.Context) error {
	// todo 
	return nil
}
```


### visible errors
visible errors will show error to client
```go
func CustomErrorWithAuto(ctx *gin.Context) error {
	log.Println("NothingToDo")
	return simple_server_runner.NewCustomError("test error")
}
```
custom error will show show：
```json
{
    "code": 500,
    "msg": "your custom error",
    "data": null
}
```
other error will show: 
```json
{
  "code": 500,
  "msg": "internal error",
  "data": null
}
```