# X-Request-ID

This package provides a functional for requests tagging with a unique ID.
Here is a few examples of usage:

#### gRPC

```go
server = grpc.NewServer(
    grpc.InTapHandle(xrequestid.New().InTapHandler),
)
```

if you don't want to use `InTapHandle` you can use `UnaryInterceptor`:

```go
server = grpc.NewServer(
    grpc.UnaryInterceptor(xrequestid.New().UnaryInterceptor),
) 
```

#### HTTP

```go
server := &http.Server{
    Handler: xrequestid.New().PopulateHTTP(http.DefaultServeMux),
}
```