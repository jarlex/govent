package actions

import "net/http"

const restType = "REST"

type RestInput struct {
    Req *http.Request
    Res *http.Response
}

type RestAction interface {
    Name() string
    Description() string
    Init() error
    Handler() func(input RestInput) error
}