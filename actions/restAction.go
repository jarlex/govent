package actions

import "net/http"

const restType = "rest"

// RestInput represents the input for a REST action.
type RestInput struct {
	Req *http.Request
	Res *http.Response
}

// RestAction defines the interface for REST actions.
type RestAction interface {
	Name() string
	Description() string
	Type() string
	Init(config map[string]interface{}) error
	Handler() func(input RestInput) error
}
