package application

import (
	"regexp"

	"github.com/Muscaw/GitFortress/internal/application/metrics"
	metricsEntity "github.com/Muscaw/GitFortress/internal/domain/metrics/entity"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/service"
	"github.com/rs/zerolog/log"

	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
)

var numberOfRepos metricsEntity.Gauge

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

func SynchronizeRepos(ignoredRepositories []*regexp.Regexp, localVcs service.LocalVCS, remoteVcs service.VCS) {
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
				log.Printf("could not clone repository %v because %+v", remoteRepo.GetFullName(), err)
			} else {
				clonedReposCount += 1
			}
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
	}
	numberOfRepos.SetInts(map[string]int{
		"local_repositories_count":        len(localRepos),
		"ignored_repositories_count":      ignoredReposCount,
		"cloned_repositories_count":       clonedReposCount,
		"synchronized_repositories_count": numberOfSynchronizedRepositories,
		"execution_count":                 1,
	})
}
