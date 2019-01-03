package types

// AuthMethod is an authentication method.
type AuthMethod string

// Settings contain the main settings of the application.
type Settings struct {
	Key        []byte              `json:"key"`
	BaseURL    string              `json:"baseURL"`
	Signup     bool                `json:"signup"`
	Defaults   UserDefaults        `json:"defaults"`
	AuthMethod AuthMethod          `json:"authMethod"`
	Branding   Branding            `json:"branding"`
	Commands   map[string][]string `json:"commands"`
	Shell      []string            `json:"shell"`
	Rules      []Rule              `json:"rules"` // TODO: use this add to cli
}

// Sorting contains a sorting order.
type Sorting struct {
	By  string `json:"by"`
	Asc bool   `json:"asc"`
}

// Branding contains the branding settings of the app.
type Branding struct {
	Name            string `json:"name"`
	DisableExternal bool   `json:"disableExternal"`
	Files           string `json:"files"`
}

// UserDefaults is a type that holds the default values
// for some fields on User.
type UserDefaults struct {
	Scope    string      `json:"scope"`
	Locale   string      `json:"locale"`
	ViewMode ViewMode    `json:"viewMode"`
	Sorting  Sorting     `json:"sorting"`
	Perm     Permissions `json:"perm"`
	Commands []string    `json:"commands"`
}
