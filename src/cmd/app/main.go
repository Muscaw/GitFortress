package main

import (
	"os"
	"regexp"
	"time"

	"github.com/Muscaw/GitFortress/config"
	"github.com/Muscaw/GitFortress/internal/application"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
	"github.com/Muscaw/GitFortress/internal/interfaces/github"
	"github.com/Muscaw/GitFortress/internal/interfaces/system_git"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
}

func main() {
	cfg := config.LoadConfig()

	client := github.GetGithubVCS(cfg.GithubToken)
	localGit := system_git.GetLocalGit(cfg.CloneFolderPath, entity.Auth{Token: cfg.GithubToken})

	var ignoredRepositoriesRegex []*regexp.Regexp
	for _, i := range cfg.IgnoreRepositories {
		ignoredRepositoriesRegex = append(ignoredRepositoriesRegex, regexp.MustCompile(i))
	}

	application.ScheduleEvery(1*time.Minute, func() {
		application.SynchronizeRepos(ignoredRepositoriesRegex, localGit, client)
	})
}
