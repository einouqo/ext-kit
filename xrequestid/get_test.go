package xrequestid

import (
	"context"
	"testing"
)

func TestGet_ok(t *testing.T) {
	ctx := context.Background()
	for _, s := range []string{
		"foo", "bar", "baz", "", "123",
	} {
		ctx = Populate(ctx, s)
		xid, ok := Get(ctx)
		if !ok {
			t.Errorf("expected ok, got !ok")
		}
		if xid != s {
			t.Errorf("expected %q, got %q", s, xid)
		}
	}
}
