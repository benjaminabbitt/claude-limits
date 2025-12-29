package api

import (
	"context"

	apierrors "github.com/benjaminabbitt/claude-limits/internal/errors"
	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all" // register all browser cookie store finders
)

// GetSessionCookieFromBrowser attempts to extract the Claude.ai session cookie from browser profiles
func GetSessionCookieFromBrowser() (string, error) {
	ctx := context.Background()

	// Use kooky to find cookies from all browsers
	cookiesSeq := kooky.TraverseCookies(ctx,
		kooky.Valid,
		kooky.DomainHasSuffix("claude.ai"),
		kooky.Name("sessionKey"),
	).OnlyCookies()

	for cookie := range cookiesSeq {
		if cookie.Value != "" {
			return cookie.Value, nil
		}
	}

	return "", apierrors.NewAuthError("browser", apierrors.ErrCookieNotFound)
}

// GetOrgIDFromBrowser attempts to extract the Claude.ai org ID from browser cookies
func GetOrgIDFromBrowser() (string, error) {
	ctx := context.Background()

	// The org ID might be in a cookie called "lastActiveOrg" or similar
	cookiesSeq := kooky.TraverseCookies(ctx,
		kooky.Valid,
		kooky.DomainHasSuffix("claude.ai"),
		kooky.Name("lastActiveOrg"),
	).OnlyCookies()

	for cookie := range cookiesSeq {
		if cookie.Value != "" {
			return cookie.Value, nil
		}
	}

	return "", apierrors.NewAuthError("browser", apierrors.ErrOrgIDNotFound)
}
