package scm

import (
	"errors"
	"net/url"
	"strings"
)

func GetDriverName(rawURL string) (string, error) {

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if s := strings.TrimSuffix(u.Host, ".com"); s != u.Host {
		return strings.ToLower(s), nil
	}

	if s := strings.TrimSuffix(u.Host, ".org"); s != u.Host {
		return strings.ToLower(s), nil
	}

	return "", errors.New("unknown Git server: " + u.Host)
}

func NewRepository(rawURL string) (Repository, error) {
	repoType, err := GetDriverName(rawURL)
	if err != nil {
		return nil, err
	}
	switch repoType {
	case "github":
		return NewGithubRepository(rawURL)
	}
	return nil, nil
}
