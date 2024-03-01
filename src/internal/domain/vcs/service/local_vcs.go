package service

import "github.com/Muscaw/GitFortress/internal/domain/vcs/entity"

type LocalVCS interface {
	VCS
	CloneRepository(repository entity.Repository) error
	SynchronizeRepository(repository entity.Repository) error
}
