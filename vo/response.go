package vo

// GeneralResponse 通用response结构
type GeneralResponse struct {
	Code    int         `json:"retCode"`
	Message string      `json:"errMsg"`
	Data    interface{} `json:"body"`
}
