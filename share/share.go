package share

// Link is the information needed to build a shareable link.
type Link struct {
	Hash   string `json:"hash" storm:"id,index"`
	Path   string `json:"path" storm:"index"`
	UserID uint   `json:"userID"`
	Expire int64  `json:"expire"`
}
