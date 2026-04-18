// Package authz provides role-hierarchy helpers shared by domain handlers.
package authz

import "github.com/luketeo/horizon/generated/oapi"

// Level returns a numeric rank for an OrgRole. Unknown roles rank 0.
func Level(role oapi.OrgRole) int {
	switch role {
	case oapi.Owner:
		return 4
	case oapi.Admin:
		return 3
	case oapi.Analyst:
		return 2
	case oapi.Viewer:
		return 1
	}
	return 0
}

// HasRole reports whether actual meets or exceeds required.
func HasRole(actual, required oapi.OrgRole) bool {
	return Level(actual) >= Level(required)
}
