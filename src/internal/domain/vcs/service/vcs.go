package service

import "github.com/Muscaw/GitFortress/internal/domain/vcs/entity"

type VCS interface {
	ListOwnedRepositories() ([]entity.Repository, error)
}

type RemoteAuthenticationProvider interface {
}
