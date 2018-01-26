package user

import (
	"github.com/ales6164/go-cms/project"
)

// namespace is email
type User struct {
	Hash      []byte             `datastore:"hash,noindex" json:"-"`
	Email     string             `datastore:"email" json:"email"`
	FirstName string             `datastore:"firstName" json:"firstName"`
	LastName  string             `datastore:"lastName" json:"lastName"`
	Photo     string             `datastore:"photo,noindex" json:"photo"`
	Projects  []*project.Project `datastore:"-" json:"projects"`
}
