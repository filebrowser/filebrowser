package settings

// ClamAV contains upload antivirus scanning settings.
type ClamAV struct {
	Enabled   bool   `json:"enabled"`
	URL       string `json:"url"`
	ScanDepth int    `json:"scanDepth"`
}
