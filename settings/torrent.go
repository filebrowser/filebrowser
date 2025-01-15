package settings

type Torrent struct {
	TrackersListUrl  string `json:"trackersListUrl"`
	QbUrl            string `json:"qbUrl"`
	QbUsername       string `json:"qbUsername"`
	QbPassword       string `json:"qbPassword"`
	AccountId        string `json:"accountId"`
	AccountKeyId     string `json:"accountKeyId"`
	AccountKeySecret string `json:"accountKeySecret"`
	Bucket           string `json:"bucket"`
	Domain           string `json:"domain"`
}
