package github

import (
	"context"

	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/service"
	"github.com/google/go-github/v58/github"
)

type githubVCS struct {
	client *github.Client
}

func (v *githubVCS) ListOwnedRepositories() ([]entity.Repository, error) {
	var allRepos []entity.Repository
	options := &github.RepositoryListByAuthenticatedUserOptions{Affiliation: "owner"}
	for {
		repos, resp, err := v.client.Repositories.ListByAuthenticatedUser(context.Background(), options)
		if err != nil {
			return nil, err
		}
		for _, r := range repos {
			allRepos = append(allRepos, githubRepositoryToDomainRepository(r))
		}
		if resp.NextPage == 0 {
			break
		}
		options.Page = resp.NextPage
	}

	return allRepos, nil
}

func GetGithubVCS(githubUrl string, githubToken string) (service.VCS, error) {
	client, err := getGithubClient(githubUrl, githubToken)
	if err != nil {
		return nil, err
	}
	return &githubVCS{client: client}, nil
}

func getGithubClient(githubUrl string, githubToken string) (*github.Client, error) {
	return github.NewClient(nil).WithAuthToken(githubToken).WithEnterpriseURLs(githubUrl, githubUrl)
}

func githubRepositoryToDomainRepository(repo *github.Repository) entity.Repository {
	return entity.Repository{
		OwnerName:      entity.OwnerName{Name: *repo.Owner.Login},
		RepositoryName: entity.RepositoryName{Name: *repo.Name},
		Remote:         entity.Remote{Name: "origin", HttpUrl: *repo.CloneURL},
	}
}
