package main

import (
	"fmt"
	"github.com/Muscaw/GitFortress/internal/application/metrics"
	"github.com/Muscaw/GitFortress/internal/interfaces/influx"
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

	metricsService := metrics.GetMetricsService()
	influxMetricHandler := influx.NewInfluxMetricsHandler()
	metricsService.RegisterHandler(influxMetricHandler)
	metricsService.Start()

	client := github.GetGithubVCS(cfg.GithubToken)
	localGit := system_git.GetLocalGit(cfg.CloneFolderPath, entity.Auth{Token: cfg.GithubToken})

	var ignoredRepositoriesRegex []*regexp.Regexp
	for _, i := range cfg.IgnoreRepositoriesRegex {
		ignoredRepositoriesRegex = append(ignoredRepositoriesRegex, regexp.MustCompile(i))
	}
	delay, err := time.ParseDuration(cfg.SyncDelay)
	if err != nil {
		panic(fmt.Errorf("could not parse configuration sync_delay value %v", cfg.SyncDelay))
	}
	if delay.Seconds() <= 0 {
		panic(fmt.Errorf("sync_delay must be a positive duration strictly superior to 0: %v", cfg.SyncDelay))
	}
	application.ScheduleEvery(delay, func() {
		application.SynchronizeRepos(ignoredRepositoriesRegex, localGit, client)
	})
}
