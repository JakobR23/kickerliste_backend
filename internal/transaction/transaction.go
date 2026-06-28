// Package transaction defines the abstraction a service depends on to run a
// unit of work inside a database transaction, without coupling to the concrete
// database wrapper. *database.DB satisfies Runner, so it can be injected
// wherever a Runner is expected.
package transaction

import "context"

// Runner runs fn inside a database transaction, committing if fn returns nil and
// rolling back otherwise. The transaction is propagated through ctx so that
// repositories invoked within fn automatically participate in it.
type Runner interface {
	Transactional(ctx context.Context, fn func(context.Context) error) error
}
