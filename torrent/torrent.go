package torrent

type CreateBody struct {
	Announces []string `json:"announces"`
	Comment   string   `json:"comment"`
	Date      bool     `json:"date"`
	Name      string   `json:"name"`
	PieceLen  int      `json:"pieceLen"`
	Private   bool     `json:"private"`
	Source    string   `json:"source"`
	WebSeeds  []string `json:"webSeeds"`
}

type Torrent struct {
	Path string `json:"Path"`
}

// Link is the information needed to build a shareable link.
// type Torrent struct {
// 	Hash         string `json:"hash" storm:"id,index"`
// 	Path         string `json:"path" storm:"index"`
// 	UserID       uint   `json:"userID"`
// 	Expire       int64  `json:"expire"`
// 	PasswordHash string `json:"password_hash,omitempty"`
// 	// Token is a random value that will only be set when PasswordHash is set. It is
// 	// URL-Safe and is used to download links in password-protected shares via a
// 	// query arg.
// 	Token string `json:"token,omitempty"`
// }
//
