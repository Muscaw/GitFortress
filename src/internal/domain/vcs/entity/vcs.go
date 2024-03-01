package entity

import (
	"fmt"
	"strings"
)

type OwnerName struct {
	Name string
}

type RepositoryName struct {
	Name string
}

type Remote struct {
	Name    string
	HttpUrl string
}

func NewRemote(name string, httpUrl string) (Remote, error) {
	if strings.HasPrefix(httpUrl, "http") || strings.HasPrefix(httpUrl, "https") {
		return Remote{Name: name, HttpUrl: httpUrl}, nil
	} else {
		return Remote{}, fmt.Errorf("remote httpUrl must contain a HTTP/HTTPS url. Received %v", httpUrl)
	}
}

type Repository struct {
	OwnerName      OwnerName
	RepositoryName RepositoryName
	Remote         Remote
}

func (r Repository) GetFullName() string {
	return strings.Join([]string{r.OwnerName.Name, r.RepositoryName.Name}, "/")
}

func IsEqual(r1, r2 Repository) bool {
	return r1.OwnerName == r2.OwnerName && r1.RepositoryName == r2.RepositoryName
}
