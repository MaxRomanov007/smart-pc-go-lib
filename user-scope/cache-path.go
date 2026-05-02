package userScope

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const AppName = "smart-pc"

type CachePath string

func NewCachePath(path string) (CachePath, error) {
	const op = "user-scope.NewCachePath"

	ucd, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("%s: can not get user cache dir: %w", op, err)
	}

	return CachePath(filepath.Join(ucd, AppName, path)), nil
}

func (p *CachePath) SetValue(s string) error {
	cp, err := NewCachePath(s)
	if err != nil {
		return err
	}

	*p = cp
	return nil
}

func (p *CachePath) UnmarshalYAML(value *yaml.Node) error {
	const op = "user-scope.UnmarshalYAML"

	var str string
	if err := value.Decode(&str); err != nil {
		return fmt.Errorf("%s: failed to decode value to string: %w", op, err)
	}
	cp, err := NewCachePath(str)
	if err != nil {
		return fmt.Errorf("%s: failed to create new cache path: %w", op, err)
	}
	*p = cp

	return nil
}

func (p *CachePath) MarshalYAML() (interface{}, error) {
	const op = "user-scope.MarshalYAML"

	ucd, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("%s: can not get user cache dir: %w", op, err)
	}

	return strings.TrimPrefix(ucd, filepath.Join(ucd, AppName, "")), nil
}
