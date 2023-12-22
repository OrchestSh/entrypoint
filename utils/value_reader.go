package utils

import (
	"os"
	"strconv"
)

// valueReader defines an interface with two methods. readValue method will return the corresponding value for the given argument. getDict will return the full map from which the valueReader reads values
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

// GetValueReader returns a valueReader from which key-value pairs can be read.
// If ORCHESTSH_USE_DOCKER_SECRETS environment variable is true and ORCHESTSH_SECRETS_PATH is a valid path,
// the returned valueReader will read from values found at the files in ORCHESTSH_SECRETS_PATH.
// If ORCHESTSH_USE_FILE environment variable is a valid file, the returned valueReader will read from values found in the file
// The default valueReader returned reads values from environment variables
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
