package api

import "google.golang.org/appengine/datastore"

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


