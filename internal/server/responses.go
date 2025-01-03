package server

type DetailData interface{}

type StandardResponse struct {
	Msg    string     `json:"msg"`
	Length int        `json:"length"`
	Type   string     `json:"type"`
	Detail DetailData `json:"detail"`
}

func NewResponse(msg string, length int, t string, detail DetailData) StandardResponse {
	return StandardResponse{
		Msg:    msg,
		Length: length,
		Type:   t,
		Detail: detail,
	}
}
