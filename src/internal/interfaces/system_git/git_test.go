package system_git

import (
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
)

func Test_isDir_is_directory(t *testing.T) {
	dirName, err := os.MkdirTemp("", "test")
	if err != nil {
		t.FailNow()
	}
	defer os.Remove(dirName)

	isADirectory, err := isDir(dirName)

	if err != nil {
		t.Fatal("could not execute isDir successfully")
	}

	if !isADirectory {
		t.Fatal("created directory is not identified as one")
	}
}

func Test_isDir_is_a_file(t *testing.T) {
	dirName, err := os.MkdirTemp("", "test")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(dirName)
	os.WriteFile(dirName+"/test", []byte("hello world"), 0644)

	isADirectory, err := isDir(dirName + "/test")
	if err != nil {
		t.Fatal("could not execute isDir successfully")
	}

	if isADirectory {
		t.Fatal("created file is not a directory")
	}
}

func Test_CloneListAndSynchronizeRepositoriesIntegration(t *testing.T) {
	dirName, err := os.MkdirTemp("", "test")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(dirName)

	// Create some git repos
	localGit := GetLocalGit(dirName, entity.Auth{Token: os.Getenv("GITHUB_TOKEN")})
	repos, err := localGit.ListOwnedRepositories()
	if err != nil {
		t.Fatal("could not list owned repositories")
	}

	if len(repos) != 0 {
		t.Fatal("owned repos should be empty in new folder")
	}

	gitFortressRepo := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "Muscaw"},
		RepositoryName: entity.RepositoryName{Name: "GitFortress"},
		Remote:         entity.Remote{Name: "origin", HttpUrl: "https://github.com/Muscaw/Gitfortress"},
	}
	err = localGit.CloneRepository(gitFortressRepo)
	if err != nil {
		t.Fatalf("could not clone repository: %v", err)
	}

	err = localGit.CloneRepository(
		entity.Repository{
			OwnerName:      entity.OwnerName{Name: "Muscaw"},
			RepositoryName: entity.RepositoryName{Name: "gitea-github-sync"},
			Remote:         entity.Remote{Name: "origin", HttpUrl: "https://github.com/Muscaw/gitea-github-sync"},
		})

	if err != nil {
		t.Fatalf("could not clone repository: %v", err)
	}

	repos, err = localGit.ListOwnedRepositories()
	if err != nil {
		t.Fatalf("could not list owned repositories: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repositories, found %v", len(repos))
	}

	localGit.SynchronizeRepository(gitFortressRepo)
}

func Test_SynchronizeRepository_local_repository_has_references_not_present_on_remote(t *testing.T) {
	dirName, err := os.MkdirTemp("", "test")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(dirName)

	// Create some git repos
	localGit := GetLocalGit(dirName, entity.Auth{Token: os.Getenv("GITHUB_TOKEN")})
	gitFortressRepo := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "Muscaw"},
		RepositoryName: entity.RepositoryName{Name: "GitFortress"},
		Remote:         entity.Remote{Name: "origin", HttpUrl: "https://github.com/Muscaw/Gitfortress"},
	}
	err = localGit.CloneRepository(gitFortressRepo)
	if err != nil {
		t.Fatalf("could not clone repository: %v", err)
	}

	repoPath := path.Join(dirName, "GitFortress")
	createTagCmd := exec.Command("git", "tag", "some-non-existing-tag")
	createTagCmd.Dir = repoPath
	if err := createTagCmd.Run(); err != nil {
		t.Fatalf("could not create test tag on repo: %v", err)
	}

	listTagCmd := exec.Command("git", "tag")
	listTagCmd.Dir = repoPath
	output, err := listTagCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("could not list tags on test repo: %v", err)
	}
	if !strings.Contains(string(output), "some-non-existing-tag") {
		t.Fatalf("repository should contain tag 'some-non-existing-tag', got %v", string(output))
	}

	localGit.SynchronizeRepository(gitFortressRepo)

	listTagCmd = exec.Command("git", "tag")
	listTagCmd.Dir = repoPath
	output, err = listTagCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("could not list tags on test repo: %v", err)
	}
	if strings.Contains(string(output), "some-non-existing-tag") {
		t.Fatalf("repository should not contain tag 'some-non-existing-tag', got %v", string(output))
	}
}

func Test_ListReposInNonExistingFolder(t *testing.T) {
	localGit := GetLocalGit("/non-existing-folder", entity.Auth{Token: "not-important"})
	_, err := localGit.ListOwnedRepositories()
	if err == nil {
		t.Fatal("should return err when folder does not exist")
	}
	if err.Error() != "could not list all possible repos from folder /non-existing-folder: open .: no such file or directory" {
		t.Fatalf("error does not match expected: %v", err)
	}
}
