package model

// Used in thc client TUI
type Context struct {
	Cmd      string
	Target   string
	From     string
	Limit    int
	Resolver string
	// JsonOutput is a flag that determines whether the output should be in JSON format.
	JsonOutput bool
	// Latency is a flag that outputs only stats of a measurement
	Latency bool
	// CI flag is used to determine whether the output should be in a format that is easy to parse by a CI tool
	CI bool
}
