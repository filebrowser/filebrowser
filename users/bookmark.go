package users

// Bookmark represents a bookmark the user has set
type Bookmark struct {
	Name string `json:"name"`
	Path string `json:"path"`
}
