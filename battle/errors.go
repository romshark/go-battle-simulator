package battle

import "github.com/pkg/errors"

// ErrMissed is an error that's returned by Attack when a figher misses
var ErrMissed = errors.New("missed")

// ErrDodged is an error that's returned by TakeDamage when a figher dodges
var ErrDodged = errors.New("dodged")

// ErrNoMoreOpponents is an error that's returned by Battlefield.FindOpponent
// when no more opponents are left
var ErrNoMoreOpponents = errors.New("no more opponents left")
