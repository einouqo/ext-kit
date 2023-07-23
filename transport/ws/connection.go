package ws

type Connection interface {
	SetReadLimit(limit int64)
	EnableWriteCompression(enable bool)
	SetCompressionLevel(level int) error
}
