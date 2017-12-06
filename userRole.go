package cms

import "errors"

type Rules map[Scope]Role
type Role string
type Scope string

var Ranks = map[Role]int{
	Guest:      0,
	Subscriber: 1,
	Editor:     2,
	Admin:      3,
}

var scopes = []Scope{Read, Write, Add, Edit, Delete}

var (
	Guest      Role = "guest"
	Subscriber Role = "subscriber" // Default user role
	Editor     Role = "editor"
	Admin      Role = "admin"

	Read   Scope = "read"
	Write  Scope = "write"
	Add    Scope = "add"
	Edit   Scope = "edit"
	Delete Scope = "delete"
)

func (c Context) HasScope(e *Entity, scope Scope) bool {
	//log.Debugf(c.Context, "HasScope: Entity %s, Scope %v", e.Name, scope)

	if role, ok := e.Rules[scope]; ok {
		return c.Rank >= Ranks[role]
	}

	return false
}

var ErrForbidden = errors.New("action forbidden")