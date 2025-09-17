package utils

import (
	log "github.com/sirupsen/logrus"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadLines(t *testing.T) {
	expected := []string{"1", "2", "a", "b", "c d f", "1 4 5"}
	res, err := readLines("testdata/somelines")
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestAppendFiles(t *testing.T) {
	expected := []string{"0", "1", "2", "3", "4", "5", "6"}
	tmpCopy, err := os.CreateTemp("", "odoo_cfg")
	assert.NoError(t, err)
	defer func() {
		if err := os.Remove(tmpCopy.Name()); err != nil {
			log.Errorf("Failed to remove temporary file: %v", err)
		}
	}()

	content, err := os.ReadFile("testdata/odoo_cfg")
	assert.NoError(t, err)

	_, err = tmpCopy.Write(content)
	assert.NoError(t, err)

	err = appendFiles(tmpCopy.Name(), "testdata/odoo.d")
	assert.NoError(t, err)
	res, err := readLines(tmpCopy.Name())
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestReadFilePairs(t *testing.T) {
	f, err := os.CreateTemp("", "file")
	if err != nil {
		t.Error(err)
	}

	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			log.Errorf("Failed to remove temporary file: %v", err)
		}
	}()

	contentString := "user=root\npassword=12345"

	if _, err := f.WriteString(contentString); err != nil {
		t.Error(err)
	}

	secrets, err := readFilePairs(f.Name())
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
