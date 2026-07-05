package hollow // want `no corresponding internal/domain`

// Command is the entry point. Its internal/domain/hollow counterpart directory
// holds only a Go-suffixed file whose package clause does not parse, so it is
// not a Go package and the command must be reported.
func Command() string { return "hollow" }
