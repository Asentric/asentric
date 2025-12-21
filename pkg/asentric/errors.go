package asentric

import "errors"

var (
	ErrRuleAlreadyRegistered = errors.New("rule already registered")
	ErrInvalidRule           = errors.New("invalid rule")
)
