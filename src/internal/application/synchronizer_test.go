package application

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
)

func Test_contains(t *testing.T) {
	aRepository := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "some_owner"},
		RepositoryName: entity.RepositoryName{Name: "some_repo"},
		Remote:         entity.Remote{Name: "some_remote", HttpUrl: "https://someurl"},
	}
	sameOwnerDifferentRepository := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "some_owner"},
		RepositoryName: entity.RepositoryName{Name: "different_repo"},
		Remote: entity.Remote{
			Name:    "some_remote",
			HttpUrl: "https://differenturl",
		},
	}
	differentOwnerDifferentRepository := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "different_owner"},
		RepositoryName: entity.RepositoryName{Name: "some_repo"},
		Remote: entity.Remote{
			Name:    "some_remote",
			HttpUrl: "https://someurl",
		},
	}
	t.Run("empty slice does not match", func(t *testing.T) {
		if contains([]entity.Repository{}, aRepository) {
			t.Error("contains returns true when list is empty")
		}
	})

	t.Run("non-equivalent entries in list are unmatched", func(t *testing.T) {
		if contains([]entity.Repository{sameOwnerDifferentRepository, differentOwnerDifferentRepository}, aRepository) {
			t.Error("contains found matching object when they are different")
		}
	})

	t.Run("only matching elements found", func(t *testing.T) {
		if !contains([]entity.Repository{aRepository, aRepository}, aRepository) {
			t.Error("element is matching, but is not found")
		}
	})

	t.Run("slice contains only searched element", func(t *testing.T) {
		if !contains([]entity.Repository{aRepository}, aRepository) {
			t.Error("element is matching, but is not found")
		}
	})

	t.Run("matching elements", func(t *testing.T) {
		if !contains([]entity.Repository{aRepository, sameOwnerDifferentRepository}, aRepository) {
			t.Error("element is matching, but is not found")
		}
	})
}

func regexOrFail(regexString string, t *testing.T) *regexp.Regexp {
	r, err := regexp.Compile(regexString)
	if err != nil {
		t.Fatalf("could not compile regex %v", regexString)
	}
	return r
}

func Test_isIgnoredRepository(t *testing.T) {
	aRepository := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "some_owner"},
		RepositoryName: entity.RepositoryName{Name: "some_repo"},
		Remote:         entity.Remote{},
	}
	anotherRepository := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "different_owner"},
		RepositoryName: entity.RepositoryName{Name: "a_repo"},
		Remote:         entity.Remote{},
	}

	matchingAllRegex := regexOrFail("^.*$", t)
	matchingARepositoryRegex := regexOrFail("^some_owner/.*$", t)
	matchingAnotherRepositoryRegex := regexOrFail("^different_owner/.*$", t)

	t.Run("regex matches", func(t *testing.T) {
		if !isIgnoredRepository([]*regexp.Regexp{matchingAllRegex}, aRepository) {
			t.Error("all matching regex does not match")
		}
		if !isIgnoredRepository([]*regexp.Regexp{matchingAllRegex}, anotherRepository) {
			t.Error("all matching regex does not match")
		}
	})

	t.Run("non matching regex does not match", func(t *testing.T) {
		if isIgnoredRepository([]*regexp.Regexp{matchingARepositoryRegex}, anotherRepository) {
			t.Error("regex should not match repository")
		}
	})

	t.Run("matching regex with repository", func(t *testing.T) {
		if !isIgnoredRepository([]*regexp.Regexp{matchingARepositoryRegex}, aRepository) {
			t.Error("regex should match with repository")
		}
	})

	t.Run("two different regexes match", func(t *testing.T) {
		regexes := []*regexp.Regexp{matchingARepositoryRegex, matchingAnotherRepositoryRegex}

		if !isIgnoredRepository(regexes, aRepository) {
			t.Error("repository should match")
		}
		if !isIgnoredRepository(regexes, anotherRepository) {
			t.Error("repository should match")
		}
	})

}

type fakeLocalVcs struct {
	ownedRepos               []entity.Repository
	errorOnListOwnedRepos    error
	clonedRepositories       []entity.Repository
	errorOnCloneRepos        error
	synchronizedRepositories []entity.Repository
	errorOnSynchonizeRepos   error
}

func (f *fakeLocalVcs) ListOwnedRepositories() ([]entity.Repository, error) {
	if f.errorOnListOwnedRepos != nil {
		return []entity.Repository{}, f.errorOnListOwnedRepos
	} else {
		return f.ownedRepos, nil
	}
}

func (f *fakeLocalVcs) CloneRepository(repository entity.Repository) error {
	if f.errorOnCloneRepos != nil {
		return f.errorOnCloneRepos
	}
	f.clonedRepositories = append(f.clonedRepositories, repository)
	f.ownedRepos = append(f.ownedRepos, repository)
	return nil
}

func (f *fakeLocalVcs) SynchronizeRepository(repository entity.Repository) error {
	f.synchronizedRepositories = append(f.synchronizedRepositories, repository)
	return f.errorOnSynchonizeRepos
}

type fakeRemoteVcs struct {
	ownedRepos                 []entity.Repository
	errorWhenListingOwnedRepos error
}

func (f *fakeRemoteVcs) ListOwnedRepositories() ([]entity.Repository, error) {
	if f.errorWhenListingOwnedRepos != nil {
		return []entity.Repository{}, f.errorWhenListingOwnedRepos
	} else {
		return f.ownedRepos, nil
	}
}

func containsAll(slice1 []entity.Repository, slice2 []entity.Repository) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i := range slice1 {
		if slice1[i] != slice2[i] {
			return false
		}
	}
	return true
}

func Test_SynchronizeRepos(t *testing.T) {
	const SOME_INPUT = "some-input"
	aRepository := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "some_owner"},
		RepositoryName: entity.RepositoryName{Name: "some_repo"},
		Remote:         entity.Remote{Name: "origin", HttpUrl: "https://someurl"},
	}
	anIgnoredRepository := entity.Repository{
		OwnerName:      entity.OwnerName{Name: "some_ignored_owner"},
		RepositoryName: entity.RepositoryName{Name: "some_repo"},
		Remote:         entity.Remote{Name: "origin", HttpUrl: "https://someurl"},
	}
	ignoredRepositoryRegex := regexOrFail("^some_ignored_owner.*", t)
	t.Run("no repos are mirrored locally and no repos are ignored", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{aRepository}, errorWhenListingOwnedRepos: nil}
		localVcs := fakeLocalVcs{ownedRepos: []entity.Repository{}}
		ignoredRepositories := []*regexp.Regexp{}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		expectedClonedAndSynchronizedRepos := []entity.Repository{aRepository}
		if !containsAll(localVcs.clonedRepositories, expectedClonedAndSynchronizedRepos) {
			t.Error("repository not present in local vcs was not cloned")
		}

		if !containsAll(localVcs.synchronizedRepositories, expectedClonedAndSynchronizedRepos) {
			t.Error("repository not present in local synchronized repositories")
		}
	})

	t.Run("remote repos exist locally and are not cloned and synchronized", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{aRepository}, errorWhenListingOwnedRepos: nil}
		localVcs := fakeLocalVcs{ownedRepos: []entity.Repository{aRepository}}
		ignoredRepositories := []*regexp.Regexp{}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		if len(localVcs.clonedRepositories) != 0 {
			t.Error("locally available repository should not be cloned again")
		}

		expectedSynchronizedRepositories := []entity.Repository{aRepository}
		if !containsAll(localVcs.synchronizedRepositories, expectedSynchronizedRepositories) {
			t.Error("repository not present in local synchronized repositories")
		}
	})

	t.Run("remote repo does not exist in configured vcs, but exists locally. Synchronization occurs anyway", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{}, errorWhenListingOwnedRepos: nil}
		localVcs := fakeLocalVcs{ownedRepos: []entity.Repository{aRepository}}
		ignoredRepositories := []*regexp.Regexp{}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		if len(localVcs.clonedRepositories) != 0 {
			t.Error("locally available repository should not be cloned again")
		}

		expectedSynchronizedRepositories := []entity.Repository{aRepository}
		if !containsAll(localVcs.synchronizedRepositories, expectedSynchronizedRepositories) {
			t.Error("locally available repository must be synchronized even without remote counterpart")
		}
	})

	t.Run("ignored repository is not cloned", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{aRepository, anIgnoredRepository}, errorWhenListingOwnedRepos: nil}
		localVcs := fakeLocalVcs{ownedRepos: []entity.Repository{}}
		ignoredRepositories := []*regexp.Regexp{ignoredRepositoryRegex}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		expectedClonedAndSynchronizedRepos := []entity.Repository{aRepository}
		if !containsAll(localVcs.clonedRepositories, expectedClonedAndSynchronizedRepos) {
			t.Error("repository not present in local vcs was not cloned")
		}

		expectedSynchronizedRepositories := []entity.Repository{aRepository}
		if !containsAll(localVcs.synchronizedRepositories, expectedSynchronizedRepositories) {
			t.Error("locally available repository must be synchronized even without remote counterpart")
		}
	})

	t.Run("ignored repository is present locally and synchronized", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{anIgnoredRepository}, errorWhenListingOwnedRepos: nil}
		localVcs := fakeLocalVcs{ownedRepos: []entity.Repository{anIgnoredRepository}}
		ignoredRepositories := []*regexp.Regexp{ignoredRepositoryRegex}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		if len(localVcs.clonedRepositories) != 0 {
			t.Error("locally available repository should not be cloned again")
		}

		expectedSynchronizedRepositories := []entity.Repository{anIgnoredRepository}
		if !containsAll(localVcs.synchronizedRepositories, expectedSynchronizedRepositories) {
			t.Error("locally available repository must be synchronized even without remote counterpart")
		}
	})

	t.Run("error when listing remote vcs owned repos", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{aRepository}, errorWhenListingOwnedRepos: fmt.Errorf("could not load repos")}
		localVcs := fakeLocalVcs{}
		ignoredRepositories := []*regexp.Regexp{}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		if len(localVcs.clonedRepositories) != 0 {
			t.Error("should not clone repos if an error is return before")
		}
	})
	t.Run("error when listing local vcs owned repos", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{aRepository}}
		localVcs := fakeLocalVcs{errorOnListOwnedRepos: fmt.Errorf("could not list repos")}
		ignoredRepositories := []*regexp.Regexp{}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		if len(localVcs.clonedRepositories) != 0 {
			t.Error("should not clone repos if an error is return before")
		}

		if len(localVcs.synchronizedRepositories) != 0 {
			t.Error("should not synchronize repos if can not list repos")
		}
	})

	t.Run("error when cloning owned repos", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{aRepository}, errorWhenListingOwnedRepos: nil}
		localVcs := fakeLocalVcs{ownedRepos: []entity.Repository{}, errorOnCloneRepos: fmt.Errorf("could not list repos")}
		ignoredRepositories := []*regexp.Regexp{}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		if len(localVcs.clonedRepositories) != 0 {
			t.Error("should have cloned repos and fail on them")
		}

		if len(localVcs.synchronizedRepositories) != 0 {
			t.Error("should not sync repos if an error is return before")
		}
	})

	t.Run("error when synchronizing local vcs repo should not fail the process", func(t *testing.T) {
		remoteVcs := fakeRemoteVcs{ownedRepos: []entity.Repository{aRepository}, errorWhenListingOwnedRepos: nil}
		localVcs := fakeLocalVcs{ownedRepos: []entity.Repository{}, errorOnSynchonizeRepos: fmt.Errorf("some error")}
		ignoredRepositories := []*regexp.Regexp{}

		SynchronizeRepos(SOME_INPUT, ignoredRepositories, &localVcs, &remoteVcs)

		expectedClonedRepositories := []entity.Repository{aRepository}

		if !containsAll(localVcs.clonedRepositories, expectedClonedRepositories) {
			t.Error("should have cloned repo		SynchronizeRepos(ignoredRepositories, &localVcs, &remoteVcs)s and fail on them")
		}

		if len(localVcs.synchronizedRepositories) == 0 {
			t.Error("should go through repos even in case of error")
		}
	})
}
