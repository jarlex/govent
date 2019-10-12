package govent

import "github.com/jarlex/govent/actions"

type Protocol int

const (
    GRPC Protocol = iota
    JSON
    TEXT
)

type Trigger struct {
    Matcher map[string]string
    Actions []actions.GenericAction
}
