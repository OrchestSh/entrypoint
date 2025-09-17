package utils

import (
	"os"
	"reflect"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGetValueReader(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "testdir")
	if err != nil {
		t.Error(err)
	}
	tempFile, err := os.CreateTemp(tempDir, "tempfile")
	if err != nil {
		t.Error(err)
	}
	if _, err := tempFile.WriteString("user=root\npassword=12345"); err != nil {
		t.Error(err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			log.Errorf("Failed to clean up tempDir: %v", err)
		}
	}()

	tests := []struct {
		name         string
		expectedType valueReader
		envVars      map[string]string
	}{
		{
			name:         "envGetter",
			expectedType: &envGetter{},
			envVars:      map[string]string{"user": "root", "password": "12345"},
		},
		{
			name:         "file valueStore",
			expectedType: &valueStore{source: tempFile.Name()},
			envVars:      map[string]string{"ORCHESTSH_USE_FILE": tempFile.Name()},
		},
		{
			name:         "docker secrets valueStore",
			expectedType: &valueStore{source: tempDir},
			envVars:      map[string]string{"ORCHESTSH_USE_DOCKER_SECRETS": "true", "ORCHESTSH_SECRETS_PATH": tempDir},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.envVars {
				if err := os.Setenv(k, v); err != nil {
					t.Error(err)
				}
			}
			vr, err := GetValueReader()
			if err != nil {
				t.Error(err)
			}
			if reflect.TypeOf(vr) != reflect.TypeOf(tt.expectedType) {
				t.Errorf("expected %T, but got %T", tt.expectedType, vr)
			}
			if vr.readValue("user") != "root" {
				t.Errorf("expected %q, but got %q", "root", vr.readValue("user"))
			}
			if vr.readValue("password") != "12345" {
				t.Errorf("expected %q, but got %q", "12345", vr.readValue("password"))
			}
		})
	}
}
