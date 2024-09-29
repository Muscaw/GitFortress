package main


import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

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
	zerolog.TimeFieldFormat = "2006-01-02T15:04:05.999Z07:00"

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
	ctx, cancelFunc := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	metricsService.Start(&wg, ctx)

	delay, err := time.ParseDuration(cfg.SyncDelay)
	if err != nil {
		panic(fmt.Errorf("could not parse configuration sync_delay value %v", cfg.SyncDelay))
	}
	if delay.Seconds() <= 0 {
		panic(fmt.Errorf("sync_delay must be a positive duration strictly superior to 0: %v", cfg.SyncDelay))
	}

	stat, err := os.Stat(cfg.CloneFolderPath)
	if err != nil {
		panic(fmt.Errorf("could not get stat for clone folder path: %w", err))
	}

	if !stat.IsDir() {
		panic(fmt.Errorf("could not proceed. clone folder path is not a directory: %v", cfg.CloneFolderPath))
	}

	for _, input := range cfg.Inputs.Github {
		client, err := github.GetGithubVCS(input.TargetURL, input.APIToken)
		if err != nil {
			panic(fmt.Errorf("could not start github client %w", err))
		}
		localInputCloneFolder := path.Join(cfg.CloneFolderPath, input.Name)
		err = os.MkdirAll(localInputCloneFolder, os.ModePerm)
		if err != nil {
			panic(fmt.Errorf("could not create local clone folder for %v. path is %v", input.Name, localInputCloneFolder))
		}
		localGit := system_git.GetLocalGit(localInputCloneFolder, entity.Auth{Token: input.APIToken})

		var ignoredRepositoriesRegex []*regexp.Regexp
		for _, i := range input.IgnoreRepositoriesRegex {
			ignoredRepositoriesRegex = append(ignoredRepositoriesRegex, regexp.MustCompile(i))
		}
		go application.ScheduleEvery(&wg, &Ticker{time.NewTicker(delay)}, ctx, func() {
			application.SynchronizeRepos(ctx, input.Name, ignoredRepositoriesRegex, localGit, client)
		})
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	cancelFunc()
	log.Info().Msg("Shutting down GitFortress")
	wg.Wait()
}
