package utils

import (
	"os"
	"strconv"
)

type valueReader interface {
	readValue(string) string
	getDict() map[string]string
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
func (eg *envGetter) getDict() map[string]string {
	fullEnv := os.Environ()
	return DefaultConverter(fullEnv)
}

func (vs *valueStore) readValue(key string) string {
	return vs.dict[key]
}

func (vs *valueStore) getDict() map[string]string {
	return vs.dict
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
