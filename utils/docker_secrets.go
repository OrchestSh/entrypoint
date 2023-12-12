package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func readDockerSecrets() (map[string]string, error) {
	secrets := make(map[string]string)
	secretsPath := "/run/secrets"
	entries, err := os.ReadDir(secretsPath)
	if err != nil {
		return nil, err
	}

	vars := map[string]string{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		filePath := filepath.Join(secretsPath, e.Name())

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(content), "\n")

		for _, l := range lines {
			parts := strings.SplitN(l, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			vars[key] = value
		}
	}

	return secrets, nil
}
