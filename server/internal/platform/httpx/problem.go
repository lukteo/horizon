// Package httpx contains shared HTTP helpers for domain handlers.
package httpx

import "github.com/luketeo/horizon/generated/oapi"

// Prob constructs an RFC 9457 ProblemDetails value with the standard
// "about:blank" type URI used throughout the API.
func Prob(status int, title, detail string) oapi.ProblemDetails {
	t := "about:blank"
	return oapi.ProblemDetails{
		Type:   &t,
		Status: &status,
		Title:  &title,
		Detail: &detail,
	}
}
