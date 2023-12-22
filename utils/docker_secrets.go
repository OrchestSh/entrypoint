package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// readDockerSecrets will look for files in the specified secretsPath and return a map filled with the key-value pairs found in the files
func readDockerSecrets(secretsPath string) (map[string]string, error) {
	entries, err := os.ReadDir(secretsPath)
	if err != nil {
		return nil, err
	}

	vars := map[string]string{}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		filePath := filepath.Clean(filepath.Join(secretsPath, e.Name()))

		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(content), "\n")

		for _, l := range lines {
			if strings.TrimSpace(l) == "" {
				continue
			}
			parts := strings.SplitN(l, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			vars[key] = value
		}
	}

	return vars, nil
}
