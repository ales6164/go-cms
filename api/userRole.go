package api

import "google.golang.org/appengine/log"

type Rules map[Scope]Role
type Role int
type Scope string

var (
	SUBSCRIBER Role = 1 // Default user role
	EDITOR     Role = 2
	ADMIN      Role = 3

	WRITE  Scope = "write"
	READ   Scope = "read"
	ADD    Scope = "add"
	EDIT   Scope = "edit"
	DELETE Scope = "delete"
)

func (c Context) HasScope(e *Entity, scope Scope) bool {
	log.Debugf(c.Context, "HasScope: Entity %s, Scope %v", e.Name, scope)

	if role, ok := e.Rules[scope]; ok {
		return c.Role >= role
	}

	return false
}

func (c Context) WithScopes(scopes ...Scope) Context {
	c.scopes = map[Scope]bool{}
	for _, scope := range scopes {
		c.scopes[scope] = true
	}
	return c
}
