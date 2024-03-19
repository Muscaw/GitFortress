package application

import (
	"fmt"
	"github.com/Muscaw/GitFortress/internal/application/metrics"
	metricsEntity "github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/service"
	"github.com/rs/zerolog/log"
	"regexp"

	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
)

var runCounter metricsEntity.Counter

func init() {
	runCounter = metrics.GetMetricsService().TrackCounter("synchronization")
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

func SynchronizeRepos(ignoredRepositories []*regexp.Regexp, localVcs service.LocalVCS, remoteVcs service.VCS) {
	remoteRepos, err := remoteVcs.ListOwnedRepositories()
	if err != nil {
		panic(fmt.Errorf("could not list all owned repos: %w", err))
	}

	localRepos, err := localVcs.ListOwnedRepositories()

	if err != nil {
		panic(fmt.Errorf("could not list all owned repos: %w", err))
	}

	for _, remoteRepo := range remoteRepos {
		if isIgnoredRepository(ignoredRepositories, remoteRepo) {
			continue
		}
		if !contains(localRepos, remoteRepo) {
			log.Info().Msgf("cloning repository %v", remoteRepo.GetFullName())
			err := localVcs.CloneRepository(remoteRepo)
			if err != nil {
				log.Printf("could not clone repository %v because %+v", remoteRepo.GetFullName(), err)
			}
		}
	}

	localRepos, err = localVcs.ListOwnedRepositories()

	if err != nil {
		panic(fmt.Errorf("could not list all owned repos: %w", err))
	}

	for _, localRepo := range localRepos {
		log.Info().Msgf("pulling repository %v", localRepo.GetFullName())
		err := localVcs.SynchronizeRepository(localRepo)
		if err != nil {
			log.Error().Err(err).Msgf("could not pull repository %v", localRepo.GetFullName())
		}
	}
	runCounter.Increment("executionCount")
}
