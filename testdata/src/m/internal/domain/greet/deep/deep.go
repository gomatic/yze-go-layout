package deep

// deep is a subpackage nested beneath the greet domain package. The
// correspondence requirement applies only at the direct <cmd> segment, so this
// package needs no internal/app/commands/greet/deep counterpart and must not
// be flagged.

// Deep does nothing structural.
func Deep() string { return "deep" }
