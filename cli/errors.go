package cli

import (
	"errors"

	"github.com/tamnd/reddit-cli/reddit"
)

func isBlocked(err error) bool {
	return errors.Is(err, reddit.ErrBlocked) ||
		errors.Is(err, reddit.ErrRateLimited) ||
		errors.Is(err, reddit.ErrPrivate) ||
		errors.Is(err, reddit.ErrBanned)
}

func isNotFound(err error) bool {
	return errors.Is(err, reddit.ErrNotFound)
}
