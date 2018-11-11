package types

// AuthMethod is an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Key        []byte       `json:"key"`
	BaseURL    string       `json:"baseURL"`
	Signup     bool         `json:"signup"`
	Defaults   UserDefaults `json:"defaults"`
	AuthMethod AuthMethod   `json:"authMethod"`
}

// UserDefaults is a type that holds the default values
// for some fields on User.
type UserDefaults struct {
	Scope    string      `json:"scope"`
	Locale   string      `json:"locale"`
	ViewMode ViewMode    `json:"viewMode"`
	Perm     Permissions `json:"perm"`
}
