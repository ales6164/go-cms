package project

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
)

// namespace is set
type Project struct {
	Name        string `datastore:"name" json:"name"`
	Namespace   string `datastore:"namespace" json:"namespace"`
	Permissions string // TODO: https://github.com/ales6164/go-cms/blob/1dcfffeba1cc1bf8f0e8f025692eb205595e7955/iam.go
}

// parent: User, namespace: Project.Namespace
type ProjectAccess struct {
	Project *datastore.Key `datastore:"project" json:"project"`
	User    *datastore.Key `datastore:"user" json:"account"`
	Role    string         `datastore:"role" json:"role"` // admin, editor, viewer, ...
}

func GetUserProjects(ctx context.Context, userKey *datastore.Key) ([]*Project, error) {
	var proAccs []ProjectAccess
	_, err := datastore.NewQuery("ProjectAccess").Ancestor(userKey).GetAll(ctx, &proAccs)
	if err != nil {
		return nil, err
	}
	var proKeys []*datastore.Key
	for _, proAcc := range proAccs {
		proKeys = append(proKeys, proAcc.Project)
	}
	var pros = make([]*Project, len(proKeys))
	err = datastore.GetMulti(ctx, proKeys, pros)
	return pros, err
}

/*func GetProjectAccess(ctx Context, namespace string) (*datastore.Key, *ProjectAccess, error) {
	proAccessKey := datastore.NewKey(ctx, "ProjectAccess", namespace, 0, ctx.userKey)
	proAccess := new(ProjectAccess)
	err := datastore.Get(ctx, proAccessKey, proAccess)
	return proAccessKey, proAccess, err
}*/

func GetProject(ctx context.Context, key *datastore.Key) (*Project, error) {
	pro := new(Project)
	err := datastore.Get(ctx, key, pro)
	return pro, err
}
