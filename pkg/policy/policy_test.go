package policy

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"os"
	"strings"
	"testing"
)

const (
	OldPodDeleterRemediator = "OldPodDeleter"
	DisabledRemediatorsKey  = "disabled_remediators"
	DisabledEnvironmentVar  = "DISABLED_REMEDIATORS"
)

func init() {
	viper.AddConfigPath("../../config")
}

func TestMissingRemediatorInConfigIsNotDisabledByDefault(t *testing.T) {
	policy := RemediatorPolicy{}
	assert.Falsef(t, policy.IsDisabled(OldPodDeleterRemediator), "remediators should not be disabled by default")
}

func TestDisabledRemediatorInConfigIsDisabled(t *testing.T) {
	policy := RemediatorPolicy{DisabledRemediators: []string{OldPodDeleterRemediator}}
	assert.Truef(t, policy.IsDisabled(OldPodDeleterRemediator),
		"remediator should be disabled when added to %s", DisabledRemediatorsKey)
}

func TestIsDisabledCaseInsensitive(t *testing.T) {
	policy := RemediatorPolicy{DisabledRemediators: []string{strings.ToLower(OldPodDeleterRemediator)}}
	assert.Truef(t, policy.IsDisabled(OldPodDeleterRemediator),
		"remediator should be disabled when added to %s", DisabledRemediatorsKey)
}

func TestLoadRemediatorsFromDefaultPolicy(t *testing.T) {
	policy := LoadRemediatorPolicy()

	assert.Empty(t, policy.DisabledRemediators, "should not have any disabled remediators")
	assert.Falsef(t, policy.IsDisabled(OldPodDeleterRemediator), "remediators should not be disabled by default")
}

func TestLoadRemediatorsUsesEnvVar(t *testing.T) {
	os.Setenv(DisabledEnvironmentVar, OldPodDeleterRemediator)
	defer os.Unsetenv(DisabledEnvironmentVar)

	policy := LoadRemediatorPolicy()
	assert.Truef(t, policy.IsDisabled(OldPodDeleterRemediator),
		"remediator should be disabled when set via env variable %s", DisabledEnvironmentVar)
}
