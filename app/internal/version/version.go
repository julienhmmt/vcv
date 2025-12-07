package version

// Build information set via ldflags at compile time.
var (
	// Version is the semantic version of the application.
	Version = "1.0"
)

// Info returns version information as a structured map.
func Info() map[string]string {
	return map[string]string{
		"version": Version,
	}
}
