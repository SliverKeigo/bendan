package platform

import "errors"

// ErrUnsupported indicates that the active platform cannot perform an action.
var ErrUnsupported = errors.New("platform operation is unsupported")
