package users

type CreateTorrentBody struct {
	Announces []string `json:"announces"`
	Comment   string   `json:"comment"`
	Date      bool     `json:"date"`
	Name      string   `json:"name"`
	PieceLen  int      `json:"pieceLen"`
	Private   bool     `json:"private"`
	Source    string   `json:"source"`
	WebSeeds  []string `json:"webSeeds"`
}
