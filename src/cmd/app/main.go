package main

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/Muscaw/GitFortress/internal/application/metrics"
	"github.com/Muscaw/GitFortress/internal/interfaces/influx"
	"github.com/Muscaw/GitFortress/internal/interfaces/prometheus"
	"github.com/rs/zerolog"

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
	if cfg.InfluxDB != nil {
		influxConfig := cfg.InfluxDB
		influxMetricHandler := influx.NewInfluxMetricsHandler(influx.MetricHandlerOpts{
			InfluxDBUrl:       influxConfig.Url,
			InfluxDBAuthToken: influxConfig.AuthToken,
			InfluxDBOrg:       influxConfig.OrganizationName,
			InfluxDBBucket:    influxConfig.BucketName,
			MetricNamePrefix:  commonMetricNamePrefix,
		})
		metricsService.RegisterHandler(influxMetricHandler)
	}
	if cfg.Prometheus != nil {
		prometheusConfig := cfg.Prometheus
		prometheusMetricHandler := prometheus.NewPrometheusMetricsHandler(
			prometheus.MetricsHandlerOpts{
				ExposedPort:      prometheusConfig.ExposedPort,
				AutoConvertNames: prometheusConfig.AutoConvertNames,
				MetricPrefix:     commonMetricNamePrefix,
			},
		)
		metricsService.RegisterHandler(prometheusMetricHandler)
	}
	ctx := context.Background()
	metricsService.Start(ctx)

	delay, err := time.ParseDuration(cfg.SyncDelay)
	if err != nil {
		panic(fmt.Errorf("could not parse configuration sync_delay value %v", cfg.SyncDelay))
	}
	if delay.Seconds() <= 0 {
		panic(fmt.Errorf("sync_delay must be a positive duration strictly superior to 0: %v", cfg.SyncDelay))
	}

	for _, input := range cfg.Inputs {
		client, err := github.GetGithubVCS(input.TargetURL, input.APIToken)
		if err != nil {
			panic(fmt.Errorf("could not start github client %w", err))
		}
		localGit := system_git.GetLocalGit(cfg.CloneFolderPath, entity.Auth{Token: input.APIToken})

		var ignoredRepositoriesRegex []*regexp.Regexp
		for _, i := range input.IgnoreRepositoriesRegex {
			ignoredRepositoriesRegex = append(ignoredRepositoriesRegex, regexp.MustCompile(i))
		}
		application.ScheduleEvery(&Ticker{time.NewTicker(delay)}, ctx, func() {
			application.SynchronizeRepos(ignoredRepositoriesRegex, localGit, client)
		})
	}
}
