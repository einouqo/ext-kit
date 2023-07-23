package pb

//go:generate protoc echo.proto --go_out=. --go-grpc_out=. --go_opt=Mecho.proto=github.com/einouqo/ext-kit/test/pb --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --go-grpc_opt=Mecho.proto=github.com/einouqo/ext-kit/test/pb
