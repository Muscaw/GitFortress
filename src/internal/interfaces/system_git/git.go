package system_git

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/service"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	log "github.com/sirupsen/logrus"
)

func isDir(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func isHttpUrl(url string) bool {
	return strings.HasPrefix(url, "http") || strings.HasPrefix(url, "https")
}

func parseRepositoryName(name string) string {
	return strings.TrimSuffix(name, ".git")
}

func parseGithubHttpUrl(url string) (string, string, error) {
	parts := strings.Split(url, "/")
	ownerRepo := parts[len(parts)-2:]
	if len(ownerRepo) != 2 {
		return "", "", fmt.Errorf("could not parse Github HTTP url: %v", url)
	}
	return ownerRepo[0], parseRepositoryName(ownerRepo[1]), nil
}

func parseOwnerRepositoryNameFromRemote(remote entity.Remote) (string, string, error) {
	if strings.Contains(remote.HttpUrl, "github.com") {
		if isHttpUrl(remote.HttpUrl) {
			return parseGithubHttpUrl(remote.HttpUrl)
		} else {
			return "", "", fmt.Errorf("SSH urls are not supported %v", remote.HttpUrl)
		}
	}
	return "", "", fmt.Errorf("remote git provider unsupported %v", remote.HttpUrl)
}

func gitRemoteToDomainRemote(remote *git.Remote) (entity.Remote, error) {
	cfg := remote.Config()
	if len(cfg.URLs) != 1 {
		return entity.Remote{
			Name:    "",
			HttpUrl: "",
		}, fmt.Errorf("Remote contains too many URLs: %v", cfg.URLs)
	}
	return entity.NewRemote(cfg.Name, cfg.URLs[0])
}

type localGitVCS struct {
	cloneDirectory string
	authentication entity.Auth
}

func (l localGitVCS) ListOwnedRepositories() ([]entity.Repository, error) {
	cloneFolder := os.DirFS(l.cloneDirectory)
	possibleRepos, err := fs.ReadDir(cloneFolder, ".")
	if err != nil {
		return nil, fmt.Errorf("could not list all possible repos from folder %v: %w", cloneFolder, err)
	}

	var foundRepos []entity.Repository
	for _, folder := range possibleRepos {
		possibleRepo := filepath.Join(l.cloneDirectory, folder.Name())
		if validPath, err := isDir(possibleRepo); err == nil && validPath {
			repo, err := git.PlainOpen(possibleRepo)
			if err != nil {
				return nil, fmt.Errorf("could not open repository %v: %w", possibleRepo, err)
			}
			remotes, err := repo.Remotes()
			if err != nil {
				return nil, fmt.Errorf("could not list all remotes for repository %v: %w", possibleRepo, err)
			}

			var originRemote *entity.Remote = nil
			for _, remote := range remotes {
				domainRemote, err := gitRemoteToDomainRemote(remote)
				if err != nil {
					return nil, fmt.Errorf("could not convert remote to domain %v: %w", remote.String(), err)
				}
				if domainRemote.Name == "origin" {
					originRemote = &domainRemote
					break
				}
			}

			if originRemote == nil {
				return nil, fmt.Errorf("could not find 'origin' remote for %v", possibleRepo)
			}

			owner, repoName, err := parseOwnerRepositoryNameFromRemote(*originRemote)
			if err != nil {
				return nil, fmt.Errorf("could not parse owner and repo from remote %v: %w", originRemote.Name, err)
			}
			foundRepos = append(foundRepos, entity.Repository{
				OwnerName:      entity.OwnerName{Name: owner},
				RepositoryName: entity.RepositoryName{Name: repoName},
				Remote:         *originRemote,
			})
		}
	}
	return foundRepos, nil
}

func (l localGitVCS) CloneRepository(repository entity.Repository) error {
	_, err := git.PlainClone(l.getRepositoryPath(repository), false, &git.CloneOptions{
		URL:    repository.Remote.HttpUrl,
		Auth:   l.getAuthentication(),
		Mirror: true,
	})
	if err != nil {
		return fmt.Errorf("could not clone repository. %w", err)
	}
	return nil
}

func contains(references []*plumbing.Reference, item *plumbing.Reference) bool {
	for _, ref := range references {
		if ref.Type() == item.Type() && ref.Name() == item.Name() {
			return true
		}
	}
	return false
}

func (l localGitVCS) prune(repo *git.Repository, targetRemote entity.Remote) error {
	remote, err := repo.Remote(targetRemote.Name)
	if err != nil {
		return fmt.Errorf("could not open remote %v: %w", targetRemote.Name, err)
	}

	remoteReferences, err := remote.List(&git.ListOptions{
		Auth: l.getAuthentication(),
	})
	if err != nil {
		return fmt.Errorf("could not list remote references for %v: %w", targetRemote.Name, err)
	}

	localReferences, err := repo.References()
	if err != nil {
		return fmt.Errorf("could not list local references for %v: %w", targetRemote.Name, err)
	}
	err = localReferences.ForEach(func(reference *plumbing.Reference) error {
		if !contains(remoteReferences, reference) {
			err := repo.Storer.RemoveReference(reference.Name())
			if err != nil {
				return fmt.Errorf("could not delete reference %v for remote %v: %w", reference.String(), targetRemote.Name, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error while parsing local references for %v: %w", targetRemote.Name, err)
	}
	return nil
}

func (l localGitVCS) SynchronizeRepository(repository entity.Repository) error {
	localRepo, err := git.PlainOpen(l.getRepositoryPath(repository))
	if err != nil {
		return fmt.Errorf("could not open repository %v. %w", repository.GetFullName(), err)
	}

	err = localRepo.Fetch(&git.FetchOptions{
		Auth: l.getAuthentication(),
	})
	if err != nil {
		if errors.Is(err, git.NoErrAlreadyUpToDate) {
			log.Infof("repository %v is already up to date", repository.GetFullName())
		} else {
			return fmt.Errorf("could not fetch repository %v: %w", repository.GetFullName(), err)
		}
	}

	err = l.prune(localRepo, repository.Remote)
	if err != nil {
		return fmt.Errorf("could not prune repository %v: %w", repository.GetFullName(), err)
	}

	return nil
}

func GetLocalGit(cloneDirectory string, remoteAuthentication entity.Auth) service.LocalVCS {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if cloneDirectory == "~" {
		cloneDirectory = dir
	} else if strings.HasPrefix(cloneDirectory, "~/") {
		cloneDirectory = filepath.Join(dir, cloneDirectory[2:])
	}
	return &localGitVCS{cloneDirectory: cloneDirectory, authentication: remoteAuthentication}
}

func (l localGitVCS) getAuthentication() *http.BasicAuth {
	return &http.BasicAuth{Username: "git", Password: l.authentication.Token}
}

func (l localGitVCS) getRepositoryPath(repo entity.Repository) string {
	return filepath.Join(l.cloneDirectory, repo.RepositoryName.Name)
}
