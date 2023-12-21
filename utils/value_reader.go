package utils

import (
	"os"
	"strconv"
)

type valueReader interface {
	readValue(string) string
}

type (
	envGetter  struct{}
	valueStore struct {
		source string
		dict   map[string]string
	}
)

func (eg *envGetter) readValue(key string) string {
	return os.Getenv(key)
}

func (fg *valueStore) readValue(key string) string {
	return fg.dict[key]
}

func (vs *valueStore) updateDict(f func(string) (map[string]string, error)) error {
	keyValueMap, err := f(vs.source)
	if err != nil {
		return err
	}
	vs.dict = keyValueMap
	return nil
}

func GetValueReader() (valueReader, error) {
	useDockerSecrets := false
	if os.Getenv("ORCHESTSH_USE_DOCKER_SECRETS") != "" {
		newValue, err := strconv.ParseBool(os.Getenv("ORCHESTSH_USE_DOCKER_SECRETS"))
		if err != nil {
			return nil, err
		}
		useDockerSecrets = newValue
	}

	secretsPath := os.Getenv("ORCHESTSH_SECRETS_PATH")
	useFile := os.Getenv("ORCHESTSH_USE_FILE")

	switch {
	case useDockerSecrets && secretsPath != "":
		dockerSecretsVS := valueStore{source: secretsPath}
		if err := dockerSecretsVS.updateDict(readDockerSecrets); err != nil {
			return nil, err
		}
		return &dockerSecretsVS, nil

	case useFile != "":
		fileVS := valueStore{source: useFile}
		if err := fileVS.updateDict(readFilePairs); err != nil {
			return nil, err
		}
		return &fileVS, nil
	default:
		return &envGetter{}, nil
	}
}
