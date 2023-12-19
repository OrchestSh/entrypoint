package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func readFileSecrets(file string) (map[string]string, error) {
	file = filepath.Clean(file)

	content, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")

	vars := map[string]string{}

	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		parts := strings.SplitN(l, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		vars[key] = value
	}

	return vars, nil
}
