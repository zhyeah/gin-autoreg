package data

// HTTPRequest route info
type HTTPRequest struct {
	URL    string
	Method string
	Func   string
	Auth   bool
	Author string
	Data   string
}

// RouterContext context
type RouterContext struct {
	HTTPMap map[string]*HTTPRequest
}
