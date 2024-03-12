package config

import (
	"github.com/bmatcuk/doublestar/v4"
	"path/filepath"
	"strings"

	"github.com/yoheimuta/protolint/internal/filepathutil"
)

// Directories represents the target directories.
type Directories struct {
	Exclude        []string `yaml:"exclude"`
	ExcludePattern []string `yaml:"exclude_pattern"`
}

func (d Directories) shouldSkipRule(
	displayPath string,
) bool {
	for _, exclude := range d.Exclude {
		if !strings.HasSuffix(exclude, string(filepathutil.OSPathSeparator)) {
			exclude += string(filepathutil.OSPathSeparator)
		}
		if filepathutil.HasUnixPathPrefix(displayPath, exclude) {
			return true
		}
	}
	for _, exclude := range d.ExcludePattern {
		if !strings.HasSuffix(exclude, "/**/") {
			exclude += "/**/"
		}
		isMatch, err := doublestar.Match(exclude, filepath.Dir(filepathutil.ConvertToUnixPath(displayPath))+"/")
		if err == nil && isMatch {
			return true
		}
	}
	return false
}
