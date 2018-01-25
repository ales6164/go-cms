package api

import (
	"google.golang.org/appengine/datastore"

	"net/http"
	"google.golang.org/appengine"
	"encoding/json"
	"golang.org/x/net/context"
	"github.com/gorilla/mux"
	"github.com/ales6164/go-cms/entity"
)

type Post struct {
	*entity.Meta
	Title        string         `datastore:"title,noindex" json:"title"`
	Category     *datastore.Key `datastore:"category,noindex" json:"-"`
	CategorySlug string         `datastore:"-" json:"category"`
	Slug         string         `datastore:"slug" json:"slug"`
	Photo        string         `datastore:"photo,noindex" json:"photo"`
	Summary      string         `datastore:"summary,noindex" json:"summary"`
	Body         []byte         `datastore:"body,noindex" json:"body"`
	Status       string         `datastore:"status" json:"status"`
}

type Category struct {
	Name   string         `datastore:"name,noindex" json:"name"`
	Slug   string         `datastore:"slug" json:"slug"`
	Parent *datastore.Key `datastore:"parent,noindex" json:"parent"`
}

type Comment struct {
	Post   *datastore.Key `datastore:"post,noindex" json:"post"`
	Author User           `datastore:"author,noindex" json:"author"`
	Body   []byte         `datastore:"body,noindex" json:"body"`
}

func SaveDraftPostHandler(w http.ResponseWriter, r *http.Request) {
	authenticated, ctx, _ := NewContext(r).WithBody().Authenticate()

	if !authenticated || ctx.projectAccessKey == nil {
		ctx.PrintAuthError(w)
		return
	}

	projectCtx, err := appengine.Namespace(ctx, ctx.projectNamespace)
	if err != nil {
		ctx.PrintAuthError(w)
		return
	}

	vars := mux.Vars(r)
	postId := vars["id"]

	var input Post
	err = json.Unmarshal(ctx.body, &input)
	if err != nil {
		ctx.PrintError(w, err)
		return
	}



	err = datastore.RunInTransaction(projectCtx, func(tc context.Context) error {
		// 1. check slug
		postKey := datastore.NewKey(tc, "Post", input.Slug, 0, nil)
		err = datastore.Get(tc, postKey, &datastore.PropertyList{})
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				// all is fine - create new post
				datastore.Put(tc, postKey, input)
			}
			return err
		}
		return ErrEntrySlugDouble
	}, nil)

	ctx.PrintAuth(w, nil, pro, signedToken)

}

func PublishPostHandler(w http.ResponseWriter, r *http.Request) {
	authenticated, ctx, _ := NewContext(r).WithBody().Authenticate()

	if !authenticated || ctx.projectAccessKey == nil {
		ctx.PrintAuthError(w)
		return
	}

	projectCtx, err := appengine.Namespace(ctx, ctx.projectNamespace)
	if err != nil {
		ctx.PrintAuthError(w)
		return
	}

	var input Post
	err = json.Unmarshal(ctx.body, &input)
	if err != nil {
		ctx.PrintError(w, err)
		return
	}

	err = datastore.RunInTransaction(projectCtx, func(tc context.Context) error {
		// 1. check slug
		postKey := datastore.NewKey(tc, "Post", input.Slug, 0, nil)
		err = datastore.Get(tc, postKey, &datastore.PropertyList{})
		if err != nil {
			if err == datastore.ErrNoSuchEntity {
				// all is fine - create new post
				datastore.Put(tc, postKey, input)
			}
			return err
		}
		return ErrEntrySlugDouble
	}, nil)

	ctx.PrintAuth(w, nil, pro, signedToken)

}