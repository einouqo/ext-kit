# Go ext kit

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

This project is a modern Go language extension of the popular [go-kit](https://github.com/go-kit/kit) library. It includes gRPC streams and WebSocket transport implementations, which were previously absent in `go-kit`. These additions are intended to make the library more versatile for contemporary Go developers. A few packages with small yet practical functions and types are also included in this project for convenience and ease of use.

## Motivation

While `go-kit` is a powerful toolkit for developing microservices in Go, it lacked support for some crucial features - gRPC streams and WebSocket transport. These two elements are increasingly fundamental to modern software architectures, enabling efficient, real-time data exchange.

Recognizing the need for these features, I saw an opportunity to augment `go-kit`'s capabilities by adding them. But the integration of gRPC streams and WebSocket transport wasn't just about introducing new features - it was about embracing the evolving landscape of Go. By using the modern Go features such as generics, I've aimed to deliver a more advanced, seamless, and comprehensive toolkit for developers.

## Features

- gRPC Streams: Enhanced support for client-side, server-side and bi-side streaming in Go, integrated seamlessly into `go-kit`.
- WebSocket Transport: A robust and performant WebSocket transport layer, expanding `go-kit`'s transport capabilities to enable real-time communication.
- Helper Functions & Types: Additional utilities to assist in rapidly building your Go applications.
- Modern Go Compatibility: Built with the latest Go features like generics for a more flexible and powerful development experience.

## Getting Started

Please follow these steps to get started with the project.

```bash
go get github.com/einouqo/ext-kit
```

## Usage

#### gRPC Bi-Directional Streaming

You can refer to the [tests](github.com/einouqo/ext-kit/test/transport/grpc) for more examples.

**Server:**
```go
func NewServerBinding(svc Service, opts ...kitgrpc.ServerOption) *ServerBinding {
	return &ServerBinding{
		/* ... */
		biStream: kitgrpc.NewServerBiStream[*pb.EchoRequest](
			svc.Bi,
			decodeRequest,
			encodeResponse,
			opts...,
		),
	}
}
```

**Client:**
```go
func NewClientBinding(cc *grpc.ClientConn) *ClientBinding {
	return &ClientBinding{
		/* ... */
		BiStream: kitgrpc.NewClientBiStream[*pb.EchoResponse](
			cc,
			pb.Echo_BiStream_FullMethodName,
			encodeRequest,
			decodeResponse,
		).Endpoint(),
	}
}
```

Make a call:
```go
sendC := make(chan service.EchoRequest) // send your requests to the channel in the way you want
receive, stop, err := client.BiStream(ctx, sendC)
if err != nil {
    // handle error
}
defer stop()
for {
    msg, err := receive()
    switch {
    case errors.Is(err, endpoint.StreamDone):
        return
    case err != nil:
        // handle error
    }
    // handle message
}
```

#### WebSocket
The usage is pretty close to gRPC Bi-Directional Streaming, but with WebSocket transport inside.

You can also refer to the [tests](github.com/einouqo/ext-kit/test/transport/ws) for more examples.

**Server:**
```go
func NewServerBinding(svc Service, opts ...ws.ServerOption) *ServerBinding {
	return &ServerBinding{
		/* ... */
		Stream: ws.NewServer(
			svc.Bi,
			decodeRequest,
			encodeResponse,
			closer,
			opts...,
		),
	}
}

func closer(_ context.Context, err error) (code ws.CloseCode, msg string, deadline time.Time) {
	if err != nil {
		return ws.InternalServerErrCloseCode, err.Error(), time.Now().Add(time.Second)
	}
	return ws.NormalClosureCloseCode, "", time.Now().Add(time.Second)
}
```

**Client:**
```go
func NewClientBinding(url url.URL, opts ...ws.ClientOption) *ClientBinding {
	return &ClientBinding{
		Stream: ws.NewClient(
			url,
			encodeRequest,
			decodeResponse,
			closer,
			opts...,
		).Endpoint(),
	}
}

func closer(context.Context, error) (code ws.CloseCode, msg string, deadline time.Time) {
	return ws.NormalClosureCloseCode, "", time.Now().Add(time.Second)
}
```

Make a call:
```go
sendC := make(chan service.EchoRequest) // send your requests to the channel in the way you want
receive, stop, err := client.BiStream(ctx, sendC)
if err != nil {
    // handle error
}
defer stop()
for {
    msg, err := receive()
    switch {
    case errors.Is(err, endpoint.StreamDone):
        return
    case err != nil:
        // handle error
    }
    // handle message
}
```

**Note:** while closing `sendC` channel leads to closing send direction of the stream in case of gRPC, while closing `sendC` channel leads to sending a close control message and following connection close in case of WebSocket.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details
