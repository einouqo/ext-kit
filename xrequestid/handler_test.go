package xrequestid

import (
	"net/http"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/tap"

	"github.com/einouqo/ext-kit/util"
)

func TestImplement(t *testing.T) {
	handler := New()
	var _ tap.ServerInHandle = handler.InTapHandler
	var _ grpc.UnaryServerInterceptor = handler.UnaryInterceptor
	var _ util.Middleware[http.Handler] = handler.PopulateHTTP
}
