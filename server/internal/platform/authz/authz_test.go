package authz_test

import (
	"testing"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/platform/authz"
)

func TestHasRole(t *testing.T) {
	cases := []struct {
		name             string
		actual, required oapi.OrgRole
		want             bool
	}{
		{"owner satisfies viewer", oapi.Owner, oapi.Viewer, true},
		{"owner satisfies admin", oapi.Owner, oapi.Admin, true},
		{"owner satisfies owner", oapi.Owner, oapi.Owner, true},
		{"admin does not satisfy owner", oapi.Admin, oapi.Owner, false},
		{"admin satisfies analyst", oapi.Admin, oapi.Analyst, true},
		{"analyst satisfies viewer", oapi.Analyst, oapi.Viewer, true},
		{"viewer does not satisfy analyst", oapi.Viewer, oapi.Analyst, false},
		{"viewer satisfies viewer", oapi.Viewer, oapi.Viewer, true},
		{"unknown role satisfies nothing", oapi.OrgRole("unknown"), oapi.Viewer, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := authz.HasRole(tc.actual, tc.required); got != tc.want {
				t.Errorf("HasRole(%q, %q) = %v, want %v", tc.actual, tc.required, got, tc.want)
			}
		})
	}
}
