package main

import (
	"context"
	"fmt"
	"github.com/Muscaw/GitFortress/internal/application/metrics"
	"github.com/Muscaw/GitFortress/internal/interfaces/influx"
	"github.com/Muscaw/GitFortress/internal/interfaces/prometheus"
	"github.com/rs/zerolog"
	"regexp"
	"time"

	"github.com/Muscaw/GitFortress/config"
	"github.com/Muscaw/GitFortress/internal/application"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
	"github.com/Muscaw/GitFortress/internal/interfaces/github"
	"github.com/Muscaw/GitFortress/internal/interfaces/system_git"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

type Ticker struct {
	ticker *time.Ticker
}

func (t *Ticker) C() <-chan time.Time {
	return t.ticker.C
}

func (t *Ticker) Stop() {
	t.ticker.Stop()
}

func main() {
	cfg := config.LoadConfig()

	metricsService := metrics.GetMetricsService()
	commonMetricNamePrefix := "gitfortress"
	if cfg.InfluxDBConfig != nil {
		influxConfig := cfg.InfluxDBConfig
		influxMetricHandler := influx.NewInfluxMetricsHandler(influx.MetricHandlerOpts{
			InfluxDBUrl:       influxConfig.InfluxDBUrl,
			InfluxDBAuthToken: influxConfig.InfluxDBAuthToken,
			InfluxDBOrg:       influxConfig.OrganizationName,
			InfluxDBBucket:    influxConfig.BucketName,
			MetricNamePrefix:  commonMetricNamePrefix,
		})
		metricsService.RegisterHandler(influxMetricHandler)
	}
	if cfg.PrometheusConfig != nil {
		prometheusConfig := cfg.PrometheusConfig
		prometheusMetricHandler := prometheus.NewPrometheusMetricsHandler(
			prometheus.MetricsHandlerOpts{
				ExposedPort:      prometheusConfig.PrometheusExposedPort,
				AutoConvertNames: prometheusConfig.AutoConvertNames,
				MetricPrefix:     commonMetricNamePrefix,
			},
		)
		metricsService.RegisterHandler(prometheusMetricHandler)
	}
	ctx := context.Background()
	metricsService.Start(ctx)

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
	application.ScheduleEvery(&Ticker{time.NewTicker(delay)}, ctx, func() {
		application.SynchronizeRepos(ignoredRepositoriesRegex, localGit, client)
	})
}
