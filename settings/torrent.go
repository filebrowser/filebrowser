package settings

type Torrent struct {
	TrackersListUrl string `json:"trackersListUrl"`
	QbUrl           string `json:"qbUrl"`
	QbUsername      string `json:"qbUsername"`
	QbPassword      string `json:"qbPassword"`
}
