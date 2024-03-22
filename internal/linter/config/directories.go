package config

import (
	"github.com/bmatcuk/doublestar/v4"
	"strings"

	"github.com/yoheimuta/protolint/internal/filepathutil"
)

// Directories represents the target directories.
type Directories struct {
	Exclude        []string `yaml:"exclude"`
	ExcludePattern []string `yaml:"exclude_pattern"`
}

func (d Directories) ShouldSkip(
	displayPath string,
) bool {
	if !strings.HasSuffix(displayPath, string(filepathutil.OSPathSeparator)) {
		displayPath += string(filepathutil.OSPathSeparator)
	}
	for _, exclude := range d.Exclude {
		if filepathutil.HasUnixPathPrefix(displayPath, exclude) {
			return true
		}
	}
	for _, exclude := range d.ExcludePattern {
		isMatch, err := doublestar.Match(exclude, filepathutil.ConvertToUnixPath(displayPath))
		if err == nil && isMatch {
			return true
		}
	}
	return false
}
