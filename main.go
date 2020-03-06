package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/thundersquared/gitea-release-attach/schema"
	"github.com/akamensky/argparse"
	"github.com/parnurzeal/gorequest"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
)

var gapi schema.GiteaAPI

var logLevels = [5]string{
	0: "fatal",
	1: "info",
	2: "debug",
	3: "trace",
}

func main() {
	var parser = argparse.NewParser("gitea-release-attach", "Attach files to Gitea Releases")
	var user = parser.String("u", "user", &argparse.Options{Help: "Username for accessing Gitea"})
	var pass = parser.String("p", "pass", &argparse.Options{Help: "Password for accessing Gitea"})
	var repo = parser.String("r", "repo", &argparse.Options{Help: "Repository URL", Required: true})
	var tag = parser.String("t", "tag", &argparse.Options{Help: "Release Tag", Required: true})
	var del = parser.Flag("d", "delete", &argparse.Options{Help: "Remove all attachments from existing Release"})
	var attachments *[]os.File = parser.FileList("f", "attachment", os.O_RDWR, 0600,
		&argparse.Options{Help: "File to be attached to Release", Required: true})
	var verbose = parser.FlagCounter("v", "verbose", &argparse.Options{Help: "Tool verbosity",
		Validate: func(args []string) error {
			fmt.Println(args)
			return nil
		}})

	err := parser.Parse(os.Args)
	if err != nil || *verbose > 3 {
		log.Fatal(parser.Usage(err))
	}

	l, _ := log.ParseLevel(logLevels[*verbose])
	log.SetLevel(l)

	if CheckArgs(*user, *pass, *repo, *tag, *del, *attachments) {
		_, err := CreateRelease()

		releaseId, err := GetRelease()

		log.Debug(releaseId)

		if nil == err {
			if true == *del {
				_, err = CleanAttachments(releaseId)
			}

			_, err = UploadAttachments(releaseId, *attachments)
		}
	} else {
		os.Exit(1)
	}
}

/**
 * Check if base required parameters are set
 */
func CheckArgs(user string, pass string, repo string, tag string, del bool, attachments []os.File) bool {
	log.Info("Checking arguments and parameters")

	if len(user) < 1 {
		user = os.Getenv("GITEA_USER")
	}

	if len(pass) < 1 {
		pass = os.Getenv("GITEA_PASS")
	}

	if len(user) < 1 {
		log.Fatal("Username is missing. Either set argument or environment variable.")
		return false
	}

	if len(pass) < 1 {
		log.Fatal("Password is missing. Either set argument or environment variable.")
		return false
	}

	gapi.BaseURL, _ = RepoURLGet(repo, "BASE")
	gapi.Owner, _ = RepoURLGet(repo, "OWNER")
	gapi.Project, _ = RepoURLGet(repo, "PROJECT")
	gapi.User = user
	gapi.Pass = pass
	gapi.Tag = tag

	log.Debug(gapi)

	return true
}

/**
 * Fetch parameter from provided repo URL
 */
func RepoURLGet(repo string, key string) (string, error) {
	var keys = map[string]int{
		"BASE":    1,
		"OWNER":   2,
		"PROJECT": 3,
	}

	re := regexp.MustCompile(`(?P<host>.*)?/(?P<user>.*)?/(?P<pass>.*)?`)
	matches := re.FindStringSubmatch(repo)

	if selection, ok := keys[key]; ok {
		return matches[selection], nil
	}

	return "", errors.New("key not found")
}

func BuildAPI(path string) string {
	return gapi.BaseURL + fmt.Sprintf("/api/v1/repos/%s/%s/%s", gapi.Owner, gapi.Project, path)
}

func CreateRelease() (string, error) {
	log.Info("Creating Release")

	m := map[string]interface{}{
		"body":       "",
		"draft":      false,
		"prerelease": false,
		"name":       gapi.Tag,
		"tag_name":   gapi.Tag,
	}

	mJson, _ := json.Marshal(m)
	request := gorequest.New().SetBasicAuth(gapi.User, gapi.Pass)

	url := BuildAPI("releases")
	log.Debug("POST: ", url)

	resp, body, errs := request.Post(url).
		Send(string(mJson)).
		End()

	if len(errs) > 0 {
		log.Debug(errs)
		return "", errors.New("API call errored")
	}

	if 200 != resp.StatusCode && 409 != resp.StatusCode {
		log.Debug(resp)
		return fmt.Sprintf("HTTP %d", resp.StatusCode), errors.New("API response errored")
	}

	log.Debug("Release created successfully!")

	return body, nil
}

func GetRelease() (string, error) {
	log.Info("Fetching Releases")

	request := gorequest.New().SetBasicAuth(gapi.User, gapi.Pass)

	url := BuildAPI("releases")
	log.Debug("GET: ", url)

	resp, body, errs := request.Get(url).End()

	if len(errs) > 0 {
		return "", errors.New("API call errored")
	}

	if 200 != resp.StatusCode && 409 != resp.StatusCode {
		return fmt.Sprintf("HTTP %d", resp.StatusCode), errors.New("API response errored")
	}

	log.Debug("Releases retrieved successfully!")

	var releases schema.Releases

	_ = json.Unmarshal([]byte(body), &releases)

	log.Debug("Releases parsed successfully!")

	for _, release := range releases {
		if release.Name == gapi.Tag && release.TagName == gapi.Tag {
			log.Debug("Target Release found: ", release.ID)
			return fmt.Sprintf("%d", release.ID), nil
		}
	}

	return "", errors.New("no release found")
}

func CleanAttachments(id string) (bool, error) {
	log.Info("Cleaning attachments for Release #", id)

	request := gorequest.New().SetBasicAuth(gapi.User, gapi.Pass)

	url := BuildAPI(fmt.Sprintf("releases/%s/assets", id))
	log.Debug("GET: ", url)

	resp, body, errs := request.Get(url).End()

	if len(errs) > 0 {
		log.Error("API call errored", errs)
		return false, errors.New("API call errored")
	}

	if 200 != resp.StatusCode && 409 != resp.StatusCode {
		log.Error("API response errored", resp)
		return false, errors.New("API response errored")
	}

	var attachments schema.Attachments

	_ = json.Unmarshal([]byte(body), &attachments)

	log.Debug("Attachments parsed successfully!")

	for _, attachment := range attachments {
		log.Debug("Deleting asset #", attachment.ID)

		request := gorequest.New().SetBasicAuth(gapi.User, gapi.Pass)

		url := BuildAPI(fmt.Sprintf("releases/%s/assets/%d", id, attachment.ID))
		log.Debug("DELETE: ", url)

		resp, _, errs := request.Delete(url).End()

		if len(errs) > 0 {
			log.Error("API call errored", errs)
		}

		if 200 != resp.StatusCode && 204 != resp.StatusCode && 409 != resp.StatusCode {
			log.Error("API response errored", resp)
		}
	}

	return true, nil
}

func UploadAttachments(id string, attachments []os.File) (bool, error) {
	log.Info("Uploading attachments for Release #", id)

	for _, attachment := range attachments {
		request := gorequest.New().SetBasicAuth(gapi.User, gapi.Pass)

		url := BuildAPI(fmt.Sprintf("releases/%s/assets", id))
		log.Debug("POST: ", url)

		resp, body, errs := request.Post(url).
			Type("multipart").
			SendFile(attachment.Name(), "", "attachment").
			End()

		if len(errs) > 0 {
			log.Error("API call errored", errs)
			return false, errors.New("API call errored")
		}

		if 200 != resp.StatusCode && 201 != resp.StatusCode && 409 != resp.StatusCode {
			log.Error("API response errored", resp)
			return false, errors.New("API response errored")
		}

		log.Debug("Body: ", body)
	}

	return true, nil
}
