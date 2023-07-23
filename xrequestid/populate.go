package xrequestid

import (
	"context"
)

func Populate(parent context.Context, xid XRequestID) context.Context {
	return context.WithValue(parent, xRequestIdentKey{}, xid)
}
