package config

import (
	"flag"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	initConfigOnce sync.Once
)

func TestInitConfig(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	originalEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range originalEnv {
			parts := splitEnv(env)
			os.Setenv(parts[0], parts[1])
		}
	}()

	testCases := []struct {
		name     string
		args     []string
		env      map[string]string
		expected func() *Cfg
	}{
		{
			name: "defaults_values",
			args: []string{"cmd"},
			env:  nil,
			expected: func() *Cfg {
				return &Cfg{
					RunAddress:     runAddress,
					DatabaseURI:    databaseURI,
					AccrualAddress: accrualAddress,
				}
			},
		},
		{
			name: "flags",
			args: []string{"cmd", "-a", ":9090", "-d", "flag-db", "-r", "flag-accrual"},
			env:  nil,
			expected: func() *Cfg {
				return &Cfg{
					RunAddress:     ":9090",
					DatabaseURI:    "flag-db",
					AccrualAddress: "flag-accrual",
				}
			},
		},
		{
			name: "env_vars",
			args: []string{"cmd"},
			env: map[string]string{
				"RUN_ADDRESS":            ":9999",
				"DATABASE_URI":           "env-db",
				"ACCRUAL_SYSTEM_ADDRESS": "env-accrual",
			},
			expected: func() *Cfg {
				return &Cfg{
					RunAddress:     ":9999",
					DatabaseURI:    "env-db",
					AccrualAddress: "env-accrual",
				}
			},
		},
		{
			name: "env_vars_override_flags",
			args: []string{"cmd", "-a", ":9090", "-d", "flag-db"},
			env: map[string]string{
				"RUN_ADDRESS":  ":9999",
				"DATABASE_URI": "env-db",
			},
			expected: func() *Cfg {
				return &Cfg{
					RunAddress:     ":9999",
					DatabaseURI:    "env-db",
					AccrualAddress: accrualAddress,
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initConfigOnce = sync.Once{}
			flag.CommandLine = flag.NewFlagSet(tc.args[0], flag.ExitOnError)
			os.Args = tc.args
			os.Clearenv()
			for k, v := range tc.env {
				err := os.Setenv(k, v)
				require.NoError(t, err)
			}

			conf := InitConfig()
			expected := tc.expected()

			assert.Equal(t, expected.RunAddress, conf.RunAddress)
			assert.Equal(t, expected.DatabaseURI, conf.DatabaseURI)
			assert.Equal(t, expected.AccrualAddress, conf.AccrualAddress)
		})
	}
}

func splitEnv(env string) []string {
	for i := 0; i < len(env); i++ {
		if env[i] == '=' {
			return []string{env[:i], env[i+1:]}
		}
	}
	return []string{env, ""}
}
