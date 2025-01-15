package users

type CreateTorrentBody struct {
	Announces []string `json:"announces"`
	Comment   string   `json:"comment"`
	Date      bool     `json:"date"`
	Name      string   `json:"name"`
	PieceLen  int      `json:"pieceLen"`
	Private   bool     `json:"private"`
	R2        bool     `json:"r2"`
	Source    string   `json:"source"`
	WebSeeds  []string `json:"webSeeds"`
}
