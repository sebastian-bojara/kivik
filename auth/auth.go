// Package auth provides a standard interface for user credential validation
// by a Kivik server.
package auth

import "context"

// A Handler is used by a server to validate auth credentials.
type Handler interface {
	// Validate returns true if the credentials are valid, false otherwise.
	Validate(ctx context.Context, username, password string) (ok bool, err error)
	// Roles returns the roles to which the user belongs.
	Roles(ctx context.Context, username string) (roles []string, err error)
}
