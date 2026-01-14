package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPreRunEValidation(t *testing.T) {
	cases := []struct {
		name        string
		cluster     string
		service     string
		task        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "TestNoClusterNoServiceNoTask",
			cluster:     "",
			service:     "",
			task:        "",
			expectError: false,
		},
		{
			name:        "TestClusterOnly",
			cluster:     "my-cluster",
			service:     "",
			task:        "",
			expectError: false,
		},
		{
			name:        "TestClusterAndService",
			cluster:     "my-cluster",
			service:     "my-service",
			task:        "",
			expectError: false,
		},
		{
			name:        "TestClusterAndTask",
			cluster:     "my-cluster",
			service:     "",
			task:        "my-task-id",
			expectError: false,
		},
		{
			name:        "TestTaskWithoutCluster",
			cluster:     "",
			service:     "",
			task:        "my-task-id",
			expectError: true,
			errorMsg:    "Cluster name must be specified when specifying task",
		},
		{
			name:        "TestServiceWithoutCluster",
			cluster:     "",
			service:     "my-service",
			task:        "",
			expectError: true,
			errorMsg:    "Cluster name must be specified when specifying service",
		},
		{
			name:        "TestClusterServiceAndTask",
			cluster:     "my-cluster",
			service:     "my-service",
			task:        "my-task-id",
			expectError: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rootCmd.SetArgs([]string{})

			if c.cluster != "" {
				rootCmd.PersistentFlags().Set("cluster", c.cluster)
			} else {
				rootCmd.PersistentFlags().Set("cluster", "")
			}
			if c.service != "" {
				rootCmd.PersistentFlags().Set("service", c.service)
			} else {
				rootCmd.PersistentFlags().Set("service", "")
			}
			if c.task != "" {
				rootCmd.PersistentFlags().Set("task", c.task)
			} else {
				rootCmd.PersistentFlags().Set("task", "")
			}

			err := rootCmd.PreRunE(rootCmd, []string{})

			if c.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), c.errorMsg)
			} else {
				assert.NoError(t, err)
			}

			rootCmd.PersistentFlags().Set("cluster", "")
			rootCmd.PersistentFlags().Set("service", "")
			rootCmd.PersistentFlags().Set("task", "")
		})
	}
}

func TestGetVersion(t *testing.T) {
	v := getVersion()
	assert.Contains(t, v, "Version:")
	assert.Contains(t, v, "Commit:")
	assert.Contains(t, v, "Built date:")
	assert.Contains(t, v, "Built by:")
}
