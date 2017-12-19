package cms

import (
	"errors"
	"strings"
)

// userGroup: entityName: scope
type Permissions map[string][]string // {"public":["post:read"], "editor":["post:*"], "admin":["*:*"]}
func (p Permissions) parse() permissions {
	var perms = permissions{}
	for userGroupName, entityScopeArray := range p {
		if _, ok := perms[userGroupName]; !ok {
			perms[userGroupName] = map[string]map[Scope]bool{}
		}

		for _, entityScope := range entityScopeArray {

			// split
			var splitEntityScope = strings.Split(entityScope, ":")
			if len(splitEntityScope) != 2 {
				panic(errors.New("invalid number of segments: " + entityScope + " allowed 2 separated with :"))
			}

			// is scope valid
			switch splitEntityScope[1] {
			case "read":
			case "create":
			case "update":
			case "delete":
			case "*":
				break
			default:
				panic(errors.New("invalid scope: " + splitEntityScope[1]))
			}

			if _, ok := perms[userGroupName][splitEntityScope[0]]; !ok {
				perms[userGroupName][splitEntityScope[0]] = map[Scope]bool{}
			}

			perms[userGroupName][splitEntityScope[0]][Scope(splitEntityScope[1])] = true
		}
	}

	return perms
}

// userGroup: entityName: scope: true|false
type permissions map[string]map[string]map[Scope]bool // {"public":{"post":{"read":true}}}

type Scope string

var (
	Read   Scope = "read"
	Create Scope = "create"
	Update Scope = "update"
	Delete Scope = "delete"
)

func (ctx Context) HasPermission(e *Entity, scope Scope) (Context, error) {
	if val1, ok := ctx.api.options.permissions[ctx.UserGroup()]; ok {
		if val2, ok := val1[e.Name]; ok {
			if val3, ok := val2[scope]; ok && val3 {
				return ctx, nil
			} else if val3, ok := val2["*"]; ok && val3 {
				return ctx, nil
			}
		} else if val2, ok := val1["*"]; ok {
			if val3, ok := val2[scope]; ok && val3 {
				return ctx, nil
			} else if val3, ok := val2["*"]; ok && val3 {
				return ctx, nil
			}
		}
	}

	return ctx, ErrForbidden
}

var ErrForbidden = errors.New("action forbidden")
