package utils

import (
	"os"
	"testing"
)

func TestReadFileSecrets(t *testing.T) {
	f, err := os.CreateTemp("", "file")
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(f.Name())

	contentString := "user=root\npassword=12345"

	if _, err := f.WriteString(contentString); err != nil {
		t.Error(err)
	}

	secrets, err := readFileSecrets(f.Name())
	if err != nil {
		t.Error(err)
	}

	expectedSecrets := map[string]string{"user": "root", "password": "12345"}

	if len(secrets) != len(expectedSecrets) {
		t.Errorf("expected length to be %v, but got %v", len(expectedSecrets), len(secrets))
	}

	for k, v := range secrets {
		if v != expectedSecrets[k] {
			t.Errorf("expected %v, but got %v", expectedSecrets[k], v)
		}
	}
}
