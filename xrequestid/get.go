package xrequestid

import "context"

func Get(ctx context.Context) (xid XRequestID, ok bool) {
	val := ctx.Value(xRequestIdentKey{})
	xid, ok = val.(XRequestID)
	if !ok {
		return "", false
	}
	return xid, true
}
