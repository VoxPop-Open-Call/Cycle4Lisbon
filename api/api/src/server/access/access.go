package access

import (
	"reflect"
	"strings"
)

// AuthFunc is a function that takes an entity and a resource and returns true
// if the action should be authorized.
type AuthFunc func(ent, res any) bool

// ACL contains the methods to register and evaluate authorization rules.
type ACL struct {
	rules map[string]AuthFunc
}

// New initializes a new access control list.
func New() *ACL {
	return &ACL{make(map[string]AuthFunc)}
}

// Register adds a new rule to the ACL.
// A previously registered rule matching the same criteria will be replaced.
//
// Only the underlying types of ent and res are used: registering User{} is
// equivalent to registering &User{}.
//
// The action can be a set of actions, separated by a comma:
// 'read,write,delete'.
func (acl *ACL) Register(ent any, action string, res any, f AuthFunc) {
	actions := strings.Split(action, ",")
	for _, act := range actions {
		acl.rules[ruleId(ent, act, res)] = f
	}
}

// Authorize evaluates the rule for the given parameters.
// It returns false if there is no such rule.
func (acl *ACL) Authorize(ent any, action string, res any) bool {
	auth, ok := acl.rules[ruleId(ent, action, res)]
	if !ok {
		return false
	}

	return auth(ent, res)
}

// ruleId returns a unique identifier for the action and types of ent and res.
func ruleId(ent any, action string, res any) string {
	return strings.Join([]string{
		key(ent),
		action,
		key(res),
	}, "-")
}

// key returns an identification string for the underlying type of val,
// ignoring pointers: key(&User{}) == key(User{}).
func key(val any) string {
	return strings.Replace(reflect.TypeOf(val).String(), "*", "", -1)
}
