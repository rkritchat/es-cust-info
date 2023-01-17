package common

type Resp struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}
