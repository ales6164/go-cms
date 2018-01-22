package api

import (
	"google.golang.org/appengine/datastore"
	"golang.org/x/net/context"
	"errors"
)

// namespace is set
type Project struct {
	Name      string `datastore:"name" json:"name"`
	Namespace string `datastore:"namespace" json:"namespace"`
}

// parent: User, namespace: Project.Namespace
type ProjectAccess struct {
	Project *datastore.Key `datastore:"project" json:"project"`
	User    *datastore.Key `datastore:"user" json:"account"`
	Role    string         `datastore:"role" json:"role"` // admin, editor, viewer
}

var ErrProjectAlreadyExists = errors.New("project with that name already exists")

func NewProject(ctx Context, name, namespace string) (*datastore.Key, *datastore.Key, *Project, *ProjectAccess, error) {
	proKey := datastore.NewKey(ctx, "Project", namespace, 0, nil)
	proAccessKey := datastore.NewKey(ctx, "ProjectAccess", namespace, 0, ctx.userKey)

	pro := &Project{
		Name:      name,
		Namespace: namespace,
	}

	proAccess := &ProjectAccess{
		Project: proKey,
		User:    ctx.userKey,
		Role:    "admin",
	}

	err := datastore.RunInTransaction(ctx, func(tc context.Context) error {
		err := datastore.Get(tc, proKey, &datastore.PropertyList{})

		if err != nil && err == datastore.ErrNoSuchEntity {
			_, err = datastore.Put(tc, proKey, pro)
			if err == nil {
				_, err = datastore.Put(tc, proAccessKey, proAccess)
			}
			return err
		}
		return ErrProjectAlreadyExists
	}, &datastore.TransactionOptions{
		XG:       true,
		Attempts: 2,
	})

	return proKey, proAccessKey, pro, proAccess, err
}

func GetProjects(ctx Context) ([]*Project, error) {
	var proAccs []ProjectAccess
	_, err := datastore.NewQuery("ProjectAccess").Ancestor(ctx.userKey).GetAll(ctx, &proAccs)
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

func GetProjectAccess(ctx Context, namespace string) (*datastore.Key, *ProjectAccess, error) {
	proAccessKey := datastore.NewKey(ctx, "ProjectAccess", namespace, 0, ctx.userKey)
	proAccess := new(ProjectAccess)
	err := datastore.Get(ctx, proAccessKey, proAccess)
	return proAccessKey, proAccess, err
}

func GetProject(ctx Context, key *datastore.Key) (*Project, error) {
	pro := new(Project)
	err := datastore.Get(ctx, key, pro)
	return pro, err
}
