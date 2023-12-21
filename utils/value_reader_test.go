package utils

import (
	"os"
	"testing"
)

func TestEnvGetter_readValue(t *testing.T) {
	eg := EnvGetter{}

	res, err := eg.readValue("test")
	if err != nil {
		t.Error(err)
	}
	if res != "" {
		t.Errorf("expected %q, got %q", "", res)
	}

	os.Setenv("test", "true")

	res, err = eg.readValue("test")
	if err != nil {
		t.Error(err)
	}
	if res != "true" {
		t.Errorf("expected %q, got %q", "true", res)
	}

}

func TestFileGetter_readValue(t *testing.T) {
	tempFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Error(err)
	}
	defer os.Remove(tempFile.Name())

	content := "test=true\n"
	if _, err := tempFile.WriteString(content); err != nil {
		t.Error(err)
	}

	fg := FileGetter{tempFile.Name()}

	res, err := fg.readValue("test")
	if err != nil {
		t.Error(err)
	}
	if res != "true" {
		t.Errorf("expected %q, got %q", "true", res)
	}
	res, err = fg.readValue("not_present")
	if err != nil {
		t.Error(err)
	}
	if res != "" {
		t.Errorf("expected %q, got %q", "", res)
	}
}
