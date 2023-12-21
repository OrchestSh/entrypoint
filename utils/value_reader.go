package utils

import "os"

type valueReader interface {
	readValue(string) (string, error)
}

type EnvGetter struct{}
type FileGetter struct {
	FileName string
}
type DockerSecretsGetter struct {
	SecretsPath string
}

func (eg *EnvGetter) readValue(key string) (string, error) {
	return os.Getenv(key), nil
}

func (fg *FileGetter) readValue(key string) (string, error) {
	keyValueMap, err := readFilePairs(fg.FileName)
	if err != nil {
		return "", err
	}
	return keyValueMap[key], nil
}

func (dsg *DockerSecretsGetter) readValue(key string) (string, error) {
	keyValueMap, err := readDockerSecrets(dsg.SecretsPath)
	if err != nil {
		return "", err
	}
	return keyValueMap[key], nil
}
