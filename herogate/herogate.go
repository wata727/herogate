package herogate

import (
	"regexp"

	"github.com/sirupsen/logrus"
	git "gopkg.in/src-d/go-git.v4"
)

func detectAppFromRepo() (string, string) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		logrus.Debug("Failed to open local Git repository: " + err.Error())
		return "", ""
	}

	remote, err := repo.Remote("herogate")
	if err != nil {
		logrus.Debug("Failed to load remote config: " + err.Error())
		return "", ""
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		logrus.Debug("Failed to get remote URLs. Perhaps, this is go-git bug.")
		return "", ""
	}

	codeCommitURLPattern := regexp.MustCompile("ssh://git-codecommit.(.+).amazonaws.com/v1/repos/(.+)")
	matches := codeCommitURLPattern.FindSubmatch([]byte(urls[0]))
	if len(matches) < 3 {
		logrus.WithFields(logrus.Fields{
			"URL": urls[0],
		}).Debug("Failed to match URL pattern")
		return "", ""
	}

	logrus.WithFields(logrus.Fields{
		"Region": string(matches[1][:]),
		"App":    string(matches[2][:]),
	}).Debug("Detected application from local Git repository")
	return string(matches[1][:]), string(matches[2][:])
}
