package policy

import (
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/util/runtime"
	"strings"
)

const (
	ConfigName = "remediator_policy"
	ConfigPath = "config"
)

type RemediatorPolicy struct {
	DisabledRemediators []string `mapstructure:"disabled_remediators,omitempty"`
}

func LoadRemediatorPolicy() RemediatorPolicy {
	viper.SetConfigName(ConfigName)
	viper.SetConfigType("json")
	viper.AddConfigPath(ConfigPath)
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	runtime.Must(err)

	policy := RemediatorPolicy{}
	err = viper.Unmarshal(&policy)
	runtime.Must(err)

	return policy
}

func (p RemediatorPolicy) IsDisabled(remediator string) bool {
	if p.DisabledRemediators != nil {
		for _, disabledRemediator := range p.DisabledRemediators {
			if strings.EqualFold(disabledRemediator, remediator) {
				return true
			}
		}
	}
	return false
}
