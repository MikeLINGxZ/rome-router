# simple-server-runner
a simple framework base gin that you can esay to code your web server

## What is that?
when I use gin to codding my web server,i always need to bind request and write response.I think,if we can codding web api like codding proto server,that will be easy and distinct.

In order to make codding web api easy, I write this repo.when I use it,I just need to define a struct and some method,the framework will auto bind to router use POST with JSON.

method format like:  func (ctx *gin.Context,req {struct}) (resp {struct ptr},err error)

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
### init your server and run
```go
runner := simple_server_runner.NewDefaultServerRunner(&Server{})
runner.Run()
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
runner.CustomResponse(func(ctx *gin.Context, data interface{}, err error) {
    resp := CustomResponse{}
    if err != nil {
        resp.Error = err.Error()
        ctx.JSON(http.StatusBadGateway, resp)
        return
    }
    resp.Data = data
    ctx.JSON(http.StatusOK, resp)
    return
})
```

### Empty request OR Empty response
you can use `EmptyRequest` and `EmptyResponse`,you don't have to define an empty struct yourself.
```go
func GetAge(ctx *gin.Context, req simple_server_runner.EmptyRequest) (*simple_server_runner.EmptyResponse, error) {
	if req.UserName == nil || *req.UserName == "" {
		return nil, errors.New("user name is nil")
	}
	return &GetAgeResponse{Msg: fmt.Sprintf("%s is 20 years old", req.UserName)}, nil
}

