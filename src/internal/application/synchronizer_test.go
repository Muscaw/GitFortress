package application

import (
	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
	"regexp"
	"testing"
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
