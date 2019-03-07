package legacybridge

import (
	"os"
	"strings"
)

const (
	EnvLegacyBypassValidation = "FLOGO_LEGACY_BYPASS_VALIDATION"
)

var schemaValidationEnabled= true

func init() {
	schemaValidationEnabled = IsValidationBypassEnabled()
}

func IsValidationBypassEnabled() bool {
	schemaValidationEnv := os.Getenv(EnvLegacyBypassValidation)
	if strings.EqualFold(schemaValidationEnv, "true") {
		return true
	}

	return false
}