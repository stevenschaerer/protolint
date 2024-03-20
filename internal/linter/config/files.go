package config

import (
	"github.com/bmatcuk/doublestar/v4"
	"github.com/yoheimuta/protolint/internal/filepathutil"
	"github.com/yoheimuta/protolint/internal/stringsutil"
)

// Files represents the target files.
type Files struct {
	Exclude        []string `yaml:"exclude"`
	ExcludePattern []string `yaml:"exclude_pattern"`
}

func (d Files) shouldSkipRule(
	displayPath string,
) bool {
	if stringsutil.ContainsCrossPlatformPathInSlice(displayPath, d.Exclude) {
		return true
	}
	for _, exclude := range d.ExcludePattern {
		isMatch, err := doublestar.Match(exclude, filepathutil.ConvertToUnixPath(displayPath))
		if err == nil && isMatch {
			return true
		}
	}
	return false
}
