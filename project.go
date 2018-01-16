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
	User    *datastore.Key `datastore:"account" json:"account"`
	Role    string // admin, editor, viewer
}

var ErrProjectAlreadyExists = errors.New("project with that name already exists")

func AddProject(ctx context.Context, name, namespace string) (*datastore.Key, *Project, error) {
	pro := &Project{
		Name:      name,
		Namespace: namespace,
	}
	var key *datastore.Key
	err := datastore.RunInTransaction(ctx, func(tc context.Context) error {
		key = datastore.NewKey(tc, "Project", namespace, 0, nil)
		err := datastore.Get(tc, key, nil)
		if err != nil && err == datastore.ErrNoSuchEntity {
			// no project with that name; can create one
			_, err = datastore.Put(tc, key, pro)
			return err
		}
		return ErrProjectAlreadyExists
	}, nil)

	return key, pro, err
}

func AddProjectAccess(ctx context.Context, user *datastore.Key, project *datastore.Key) (*datastore.Key, *ProjectAccess, error) {
	pro := &ProjectAccess{
		Project: project,
		User:    user,
		Role:    "admin",
	}
	key := datastore.NewKey(ctx, "ProjectAccess", project.StringID(), 0, user)

	_, err := datastore.Put(ctx, key, pro)



	return key, pro, err
}
