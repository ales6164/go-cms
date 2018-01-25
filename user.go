package api

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"github.com/dgrijalva/jwt-go"
)

// namespace is email
type User struct {
	Hash      []byte     `datastore:"hash,noindex" json:"-"`
	Email     string     `datastore:"email" json:"email"`
	FirstName string     `datastore:"firstName" json:"firstName"`
	LastName  string     `datastore:"lastName" json:"lastName"`
	Photo     string     `datastore:"photo,noindex" json:"photo"`
	Projects  []*Project `datastore:"-" json:"projects"`
}


func getProjectAccess(ctx context.Context, usrKey *datastore.Key, projectAccessEncodedKey string) (*ProjectAccess, error) {
	proKey, err := datastore.DecodeKey(projectAccessEncodedKey)
	if err != nil {
		return nil, err
	}

	proAccKey := datastore.NewKey(ctx, "ProjectAccess", proKey.StringID(), 0, usrKey)

	proAcc := new(ProjectAccess)
	err = datastore.Get(ctx, proAccKey, proAcc)
	if err != nil {
		return nil, err
	}

	return proAcc, nil
}
