package file

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/yoheimuta/protolint/internal/linter/config"
)

// ProtoSet represents a set of .proto files.
type ProtoSet struct {
	protoFiles []ProtoFile
}

// NewProtoSet creates a new ProtoSet.
func NewProtoSet(
	targetPaths []string,
	config config.ExternalConfig,
) (ProtoSet, error) {
	fs, err := collectAllProtoFilesFromArgs(targetPaths, config)
	if err != nil {
		return ProtoSet{}, err
	}
	if len(fs) == 0 {
		return ProtoSet{}, fmt.Errorf("not found protocol buffer files in %v", targetPaths)
	}

	return ProtoSet{
		protoFiles: fs,
	}, nil
}

// ProtoFiles returns proto files.
func (s ProtoSet) ProtoFiles() []ProtoFile {
	return s.protoFiles
}

type DirectoryWalker = func(string, fs.WalkDirFunc) error

func collectAllProtoFilesFromArgs(
	targetPaths []string,
	config config.ExternalConfig,
) ([]ProtoFile, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	absCwd, err := absClean(cwd)
	if err != nil {
		return nil, err
	}
	// Eval a possible symlink for the cwd to calculate the correct relative paths in the next step.
	if newPath, err := filepath.EvalSymlinks(absCwd); err == nil {
		absCwd = newPath
	}

	var fs []ProtoFile
	for _, path := range targetPaths {
		absTarget, err := absClean(path)
		if err != nil {
			return nil, err
		}

		f, err := collectAllProtoFiles(absCwd, absTarget, config, filepath.WalkDir)
		if err != nil {
			return nil, err
		}
		fs = append(fs, f...)
	}
	return fs, nil
}

func collectAllProtoFiles(
	absWorkDirPath string,
	absPath string,
	config config.ExternalConfig,
	walker DirectoryWalker,
) ([]ProtoFile, error) {
	var files []ProtoFile

	err := walker(
		absPath,
		func(path string, info fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			displayPath, err := filepath.Rel(absWorkDirPath, path)
			if err != nil {
				displayPath = path
			}
			if config.Lint.Directories.ShouldSkip(displayPath) {
				return fs.SkipDir
			}
			if filepath.Ext(path) != ".proto" {
				return nil
			}
			if config.Lint.Files.ShouldSkip(displayPath) {
				return nil
			}
			files = append(files, NewProtoFile(path, displayPath))
			return nil
		},
	)
	if err != nil && !errors.Is(err, fs.SkipDir) {
		return nil, err
	}
	return files, nil
}

// absClean returns the cleaned absolute path of the given path.
func absClean(path string) (string, error) {
	if path == "" {
		return path, nil
	}
	if !filepath.IsAbs(path) {
		return filepath.Abs(path)
	}
	return filepath.Clean(path), nil
}
