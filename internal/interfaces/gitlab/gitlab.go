package gitlab

import (
	"net/url"

	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/service"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/xanzy/go-gitlab"
)

type gitlabVCS struct {
	client *gitlab.Client
	userId int
}

func (g *gitlabVCS) ListOwnedRepositories() ([]entity.Repository, error) {
	var allRepos []entity.Repository
	var nextPageUrl *string = nil
	for {
		nextPageOption := func(req *retryablehttp.Request) error {
			if nextPageUrl != nil {
				url, err := url.Parse(*nextPageUrl)
				if err != nil {
					return err
				}
				req.URL = url
			}
			return nil
		}
		projects, resp, err := g.client.Projects.ListUserProjects(g.userId, nil, nextPageOption)
		if err != nil {
			return nil, err
		}
		for _, r := range projects {
			allRepos = append(allRepos, gitlabProjectToDomainRepository(r))
		}
		if resp.NextLink == "" {
			break
		} else {
			nextPageUrl = &resp.NextLink
		}
	}
	return allRepos, nil
}

func GetGitlabVCS(gitlabUrl string, gitlabToken string) (service.VCS, error) {
	client, err := getGitlabClient(gitlabUrl, gitlabToken)
	if err != nil {
		return nil, err
	}
	user, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, err
	}
	return &gitlabVCS{client: client, userId: user.ID}, nil
}

func getGitlabClient(gitlabUrl string, gitlabToken string) (*gitlab.Client, error) {
	return gitlab.NewClient(gitlabToken, gitlab.WithBaseURL(gitlabUrl))
}

func gitlabProjectToDomainRepository(project *gitlab.Project) entity.Repository {
	return entity.Repository{
		OwnerName:      entity.OwnerName{Name: project.Owner.Username},
		RepositoryName: entity.RepositoryName{Name: project.Name},
		Remote:         entity.Remote{Name: "origin", HttpUrl: project.ImportURL},
	}
}
