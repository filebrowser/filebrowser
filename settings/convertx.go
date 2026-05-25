package settings

// ConvertX stores the optional ConvertX integration settings.
//
// Configured is used to distinguish an intentionally disabled GUI setting from
// an old database that predates these fields. If Configured is false, optional
// environment values can still be used as initial fallback values.
type ConvertX struct {
	Configured bool   `json:"configured"`
	Enabled    bool   `json:"enabled"`
	URL        string `json:"url"`
	APIKey     string `json:"apiKey"`
	Timeout    string `json:"timeout"`
}
