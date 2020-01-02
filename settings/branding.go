package settings

// Branding contains the branding settings of the app.
type Branding struct {
	Name            string `json:"name"`
	DisableExternal bool   `json:"disableExternal"`
	Files           string `json:"files"`
	Theme           string `json:"theme"`
}
