# gin-autoreg
register and fill params route by tag automaticlly

## Boot
You can boot gin-autoreg by follow codes, while only ```Engine``` is necessary.
```go
autoRouter := autoroute.GetAutoRouter()
	autoRouter.RegisterRoute(&autoroute.AutoRouteConfig{
		BaseUrl: "/beaconApi", // base url before each 'url' you defined in controller tag.
		Engine:  g, // assign gin engine object
		ResponseHandler: func(ctx *gin.Context, exp *exception.HTTPException, data interface{}) {
      // define your response format.
			if exp == nil {
				controller.SendResponse(ctx, nil, data)
			} else if exp.Code == http.StatusInternalServerError {
				controller.SendResponse(ctx, errno.InternalServerError, nil)
			} else {
				controller.SendResponse(ctx, errno.New(&errno.Errno{Code: util.IntToStr(exp.Code), Message: exp.Message}, exp), nil)
			}
		},
		OAAuth: midwares.SmartProxyAuth(), // your pre handler
	})
```

## Demo Controller
```go
// TestController test controller
type TestController struct {
	routeGetModuleList string `httprequest:"url=/api/test/get;func=TestGet;method=GET;auth=false"`
	routeAddModule     string `httprequest:"url=/api/test/post;func=TestPost;method=POST;auth=false"`
}

// TestGet test get request
func (controller *TestController) TestGet(request *vo.TestGetRequest) (*vo.TestGetResponse, error) {
	return &vo.TestGetResponse{
		Name:  request.Name,
		Class: request.Class,
		Age:   request.Age,
		Hobby: request.Hobby,
	}, nil
}

// TestPost test post request
func (controller *TestController) TestPost(request *vo.TestPostRequest) error {
	fmt.Println(*request)
	return nil
}

func init() {
	controller.ControllerMap["TestController"] = &TestController{}
}
```

### Explaination
As you see, we register route by adding tag for the member of controller, these members should start with 'route'. Then add ```httprequest``` tag after them. The route will be register to gin automatically.

The param of ```httprequest``` are as follows:
* url: the url path of this action.
* method: the 'method' of this http request, it can be: GET, POST, PUT, DELETE.
* func: the function of this controller which will handle this request.
* auth: give a func like ```func CheckAuth(ctx *gin.Context)```, it will handle before the ```func``` executed.
