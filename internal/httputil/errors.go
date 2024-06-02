package httputil

type HttpError400 struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"status bad request"`
}

type HttpError500 struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"internal server error"`
}

type HttpError401 struct {
	Code    int    `json:"code" example:"401"`
	Message string `json:"message" example:"Unauthorized"`
}
