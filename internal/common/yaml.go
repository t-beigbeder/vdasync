package common

import (
	"gopkg.in/yaml.v3"
	"io"
	"os"
)

// YamlLoad opens the file with provided path and unmarshal its yaml content to the value val with provided address.
func YamlLoad(path string, val interface{}) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	y, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(y, val)
	if err != nil {
		return err
	}
	return nil
}

// YamlStore marshals the provided value val as yaml and stores it as file with provided path.
func YamlStore(path string, val interface{}) error {
	y, err := yaml.Marshal(val)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(y)
	if err != nil {
		return err
	}
	return nil
}
