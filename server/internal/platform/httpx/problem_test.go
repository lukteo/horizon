package httpx_test

import (
	"testing"

	"github.com/luketeo/horizon/internal/platform/httpx"
)

func TestProb(t *testing.T) {
	p := httpx.Prob(404, "Not Found", "item missing")

	if p.Status == nil || *p.Status != 404 {
		t.Errorf("Status = %v, want 404", p.Status)
	}
	if p.Title == nil || *p.Title != "Not Found" {
		t.Errorf("Title = %v, want %q", p.Title, "Not Found")
	}
	if p.Detail == nil || *p.Detail != "item missing" {
		t.Errorf("Detail = %v, want %q", p.Detail, "item missing")
	}
	if p.Type == nil || *p.Type != "about:blank" {
		t.Errorf("Type = %v, want %q", p.Type, "about:blank")
	}
}
