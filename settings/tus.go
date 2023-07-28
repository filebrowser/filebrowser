package settings

const DefaultTusChunkSize = 10 * 1024 * 1024 // 10MB
const DefaultTusRetryCount = 5

// Tus contains the tus.io settings of the app.
type Tus struct {
	ChunkSize  uint64 `json:"chunkSize"`
	RetryCount uint16 `json:"retryCount"`
}
