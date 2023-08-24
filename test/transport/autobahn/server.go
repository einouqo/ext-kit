package autobahn

import (
	"net/http"
)

type ServerBindings struct {
	C http.Handler
}

func NewServerBindings() *ServerBindings {
	return &ServerBindings{
		// TODO: implement
	}
}
