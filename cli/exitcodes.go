package cli

// Exit codes as defined in cli-spec.md
const (
	ExitSuccess       = 0 // Success
	ExitError         = 1 // General error
	ExitUsage         = 2 // Invalid usage / bad arguments
	ExitNotFound      = 3 // Plugin not found
	ExitSourceFailed  = 4 // Source resolution failed
	ExitRegistryError = 5 // Registry/filesystem error
)
