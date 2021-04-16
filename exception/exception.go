package exception

// HTTPException 基础异常
type HTTPException struct {
	Code    int
	Message string
	Err     error
}

func (exception *HTTPException) Error() string {
	return exception.Message
}

// New 创建一个新的exception
func New(code int, message string, err error) *HTTPException {
	return &HTTPException{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
