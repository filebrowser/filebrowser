package downloader

type Downloader interface {
	Download(url string, filename string, pathname string) error
	GetRatio() float64
}
