package settings

const DefaultTusChunkSize = 20 * 1024 * 1024 // 20MB
const DefaultTusParallelUploads = 3
const DefaultTusRetryCount = 3

// Tus contains the tus.io settings of the app.
type Tus struct {
	Enabled         bool   `json:"enabled"`
	ChunkSize       uint64 `json:"chunkSize"`
	ParallelUploads uint8  `json:"parallelUploads"`
	RetryCount      uint16 `json:"retryCount"`
}
