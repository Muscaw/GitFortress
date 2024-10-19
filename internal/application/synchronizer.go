package application

import (
	"context"
	"os"
	"regexp"

	"github.com/Muscaw/GitFortress/internal/application/metrics"
	metricsEntity "github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/service"
	"github.com/rs/zerolog"

	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
)

var numberOfRepos metricsEntity.Gauge
var executionCount int = 1

func init() {
	numberOfRepos = metrics.GetMetricsService().TrackGauge("synchronization_run")
}

func contains(slice []entity.Repository, repository entity.Repository) bool {
	for _, e := range slice {
		if entity.IsEqual(e, repository) {
			return true
		}
	}
	return false
}

func isIgnoredRepository(ignoredRepositories []*regexp.Regexp, repository entity.Repository) bool {
	for _, i := range ignoredRepositories {
		if i.MatchString(repository.GetFullName()) {
			return true
		}
	}
	return false
}

func SynchronizeRepos(ctx context.Context, inputName string, ignoredRepositories []*regexp.Regexp, localVcs service.LocalVCS, remoteVcs service.VCS) {
	log := zerolog.New(os.Stdout).With().Timestamp().Str("input", inputName).Logger()
	remoteRepos, err := remoteVcs.ListOwnedRepositories()
	if err != nil {
		log.Err(err).Msg("could not list all owned repos")
		return
	}

	localRepos, err := localVcs.ListOwnedRepositories()

	if err != nil {
		log.Err(err).Msg("could not list all owned repos")
		return
	}

	ignoredReposCount := 0
	clonedReposCount := 0
	for _, remoteRepo := range remoteRepos {
		if isIgnoredRepository(ignoredRepositories, remoteRepo) {
			ignoredReposCount += 1
			continue
		}
		if !contains(localRepos, remoteRepo) {
			log.Info().Msgf("cloning repository %v", remoteRepo.GetFullName())
			err := localVcs.CloneRepository(remoteRepo)
			if err != nil {
				log.Err(err).Msgf("could not clone repository %v", remoteRepo.GetFullName())
			} else {
				clonedReposCount += 1
			}
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	localRepos, err = localVcs.ListOwnedRepositories()

	if err != nil {
		log.Err(err).Msg("could not list all owned repos")
		return
	}

	numberOfSynchronizedRepositories := 0
	for _, localRepo := range localRepos {
		log.Info().Msgf("pulling repository %v", localRepo.GetFullName())
		err := localVcs.SynchronizeRepository(localRepo)
		if err != nil {
			log.Error().Err(err).Msgf("could not pull repository %v", localRepo.GetFullName())
		} else {
			numberOfSynchronizedRepositories += 1
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
	numberOfRepos.SetInts(map[string]int{
		"remote_repositories_count":       len(remoteRepos),
		"local_repositories_count":        len(localRepos),
		"ignored_repositories_count":      ignoredReposCount,
		"cloned_repositories_count":       clonedReposCount,
		"synchronized_repositories_count": numberOfSynchronizedRepositories,
		"execution_count":                 executionCount,
	})
	executionCount += 1
}
