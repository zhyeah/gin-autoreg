# gin-autoreg
register and fill params route by tag automaticlly

## 1. Boot
You can boot gin-autoreg by follow codes, while only ```Engine``` is required.
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

### Explanation
Firstly, you need to get the instance of AutoRouter by ```GetAutoRouter```, then invoke ```RegisterRoute``` to kick off the procedure.
* BaseUrl: the common url added before each of your api.
* Engine: gin engine object.
* OAAuth: give a func like ```func CheckAuth(ctx *gin.Context)```
* ResponseHandler: the respone given by this package default is like:
  ```json
  {
    "retCode": 200,
    "errMsg": "",
    "body": {}
  }
  ```
  You can give your owner format by using ```gin.Context```


## 2. Demo Controller
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

The ```vo.TestGetRequest``` defined as:
```go
type TestGetRequest struct {
	Name  string `from:"query" field:"name"`
	Class string `from:"query" default:"6-01"`
	Age   int    `from:"query"`
	Hobby string `from:"hobby" must:"false"`
}
```

And ```vo.TestPostRequest``` is as follow:
```go
type TestPostRequest struct {
	Data TestPostBody `from:"body"`
}

type TestPostBody struct {
	Name  string `json:"name"`
	Class string `json:"class"`
	Age   int    `json:"age"`
	Hobby string `json:"hobby"`
}
```

### 2.1 Explaination
#### 2.1.1 Route
As you see, we register route by adding tag for the member of controller, these members should start with 'route'. Then add ```httprequest``` tag after them. The route will be register to gin automatically.

The param of ```httprequest``` are as follows:
* url: the url path of this action.
* method: the 'method' of this http request, it can be: GET, POST, PUT, DELETE.
* func: the function of this controller which will handle this request.
* auth: when true it will execute the ```OAAuth``` method you given in 'Boot' before the ```func``` executed.

Note: ```controller.ControllerMap["TestController"] = &TestController{}``` is the command that add your controller into register list, don't forget it!

#### 2.1.2 Parameter
You can use ```*gin.Context``` as the argument for your controller, and you can also define your own struct instead. The tag applied for this are:

* field: the field name that this value stored in.
* from: 
  * query: the field value comes from url parameter.
  * path: the field value comes from url path.
  * form: the field value comes from the post form data.
  * body: the field value comes from raw body data.
  * context: the field value comes from ```gin.Context```.
* default: if this field is not required, you can give it a default value.
* must: if this field is required, assign ```true``` to it, otherwise ```false```


### 2.2 Test
#### 2.2.1 Get Method
Now let's try to query the url 'http://{host}:{port}/api/test/get?name=abc&age=18', and we get
```json
{
    "retCode": 0,
    "errMsg": "",
    "body": {
        "name": "abc",
        "class": "6-01",
        "age": 18,
        "hobby": ""
    }
}
```

And if we drop ```age``` argument, we get
```json
{
    "retCode": 400,
    "errMsg": "field 'age' val '' cannot convert to int",
    "body": null
}
```
