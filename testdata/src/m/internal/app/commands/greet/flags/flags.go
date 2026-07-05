package flags

// flags is a subpackage nested beneath the greet command. The correspondence
// requirement applies only at the direct <cmd> segment, so this package needs
// no internal/domain/greet/flags counterpart and must not be flagged.

// Flag does nothing structural.
func Flag() string { return "flags" }
