package codeowner

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hmarr/codeowners"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ruleset codeowners.Ruleset

func getCodeOwnersPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir = filepath.Clean(dir)
	return filepath.Join(dir, "CODEOWNERS"), nil
}

func LoadOwners(path string) ([]string, error) {
	if ruleset == nil {
		codeownersPath, err := getCodeOwnersPath()
		if err != nil {
			return nil, err
		}

		b, err := os.ReadFile(codeownersPath)
		if err != nil {
			return nil, err
		}

		buf := bytes.NewBuffer(b)
		ruleset, err = codeowners.ParseFile(buf)
		if err != nil {
			return nil, err
		}
	}

	name := filepath.Base(path)
	match, err := ruleset.Match(name)
	if err != nil {
		return nil, err
	}

	r, err := regexp.Compile(`^(report to\:\ +).*`)
	if err != nil {
		return nil, err
	}

	if !r.MatchString(match.Comment) {
		return nil, errors.New("missing report comment")
	}

	owners := strings.Split(strings.TrimSpace(strings.TrimPrefix(match.Comment, "report to: ")), " ")
	for i := 0; i < len(owners); i++ {
		owners[i] = strings.TrimPrefix(owners[i], "<@")
		owners[i] = strings.TrimSuffix(owners[i], ">")
		owners[i] = fmt.Sprintf("<@%s>", owners[i])
	}

	return owners, nil
}
