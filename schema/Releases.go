package schema

import "time"

type Releases []struct {
	ID              int           `json:"id"`
	TagName         string        `json:"tag_name"`
	TargetCommitish string        `json:"target_commitish"`
	Name            string        `json:"name"`
	Body            string        `json:"body"`
	URL             string        `json:"url"`
	TarballURL      string        `json:"tarball_url"`
	ZipballURL      string        `json:"zipball_url"`
	Draft           bool          `json:"draft"`
	Prerelease      bool          `json:"prerelease"`
	CreatedAt       time.Time     `json:"created_at"`
	PublishedAt     time.Time     `json:"published_at"`
	Author          Author        `json:"author"`
	Assets          []interface{} `json:"assets"`
}

type Author struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	FullName  string    `json:"full_name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url"`
	Language  string    `json:"language"`
	IsAdmin   bool      `json:"is_admin"`
	LastLogin time.Time `json:"last_login"`
	Created   time.Time `json:"created"`
	Username  string    `json:"username"`
}
