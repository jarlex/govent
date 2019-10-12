package actions

import (
    "fmt"
    "plugin"
)

type GenericAction interface {
    Name() string
    Description() string
    Type() string
    Init() error
    Handler() func(interface{}) error
}

func New(path string) (GenericAction, error) {
    so, err := plugin.Open(path)
    if err != nil {
        return nil, err
    }
    
    act, err := so.Lookup("Plugin")
    if err != nil {
        return nil, err
    }
    return act.(GenericAction), nil
}

func Execute(ac GenericAction, input interface{}) error {
    switch ac.Type() {
    case restType:
        return ac.(RestAction).Handler()(input.(RestInput))
    default:
        return fmt.Errorf("action type %s not supported",ac.Type())
    }
}
