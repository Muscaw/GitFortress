package service

import (
	"fmt"
	"github.com/Muscaw/GitFortress/internal/domain/vcs/entity"
)

var ListOwnedRepositoriesTransientError = fmt.Errorf("could not list owned repos. please retry")

type VCS interface {
	ListOwnedRepositories() ([]entity.Repository, error)
}

type RemoteAuthenticationProvider interface {
}
