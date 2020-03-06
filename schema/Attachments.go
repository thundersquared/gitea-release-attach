package schema

import "time"

type Attachments []struct {
	ID                 int       `json:"id"`
	Name               string    `json:"name"`
	Size               int       `json:"size"`
	DownloadCount      int       `json:"download_count"`
	CreatedAt          time.Time `json:"created_at"`
	UUID               string    `json:"uuid"`
	BrowserDownloadURL string    `json:"browser_download_url"`
}
