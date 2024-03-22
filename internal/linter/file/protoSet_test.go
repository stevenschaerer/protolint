package file

import (
	"errors"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yoheimuta/protolint/internal/linter/config"

	"github.com/yoheimuta/protolint/internal/setting_test"
)

func TestNewProtoSet(t *testing.T) {
	tests := []struct {
		name             string
		inputTargetPaths []string
		wantProtoFiles   []ProtoFile
		wantExistErr     bool
	}{
		{
			name: "innerdir3 includes no files",
			inputTargetPaths: []string{
				setting_test.TestDataPath("testdir", "innerdir3"),
			},
			wantExistErr: true,
		},
		{
			name: "innerdir2 includes no proto files",
			inputTargetPaths: []string{
				setting_test.TestDataPath("testdir", "innerdir2"),
			},
			wantExistErr: true,
		},
		{
			name: "innerdir includes a proto file",
			inputTargetPaths: []string{
				setting_test.TestDataPath("testdir", "innerdir"),
			},
			wantProtoFiles: []ProtoFile{
				NewProtoFile(
					filepath.Join(setting_test.TestDataPath("testdir", "innerdir"), "/testinner.proto"),
					"../../../_testdata/testdir/innerdir/testinner.proto",
				),
			},
		},
		{
			name: "testdir includes proto files and inner dirs",
			inputTargetPaths: []string{
				setting_test.TestDataPath("testdir"),
			},
			wantProtoFiles: []ProtoFile{
				NewProtoFile(
					filepath.Join(setting_test.TestDataPath("testdir", "innerdir"), "/testinner.proto"),
					"../../../_testdata/testdir/innerdir/testinner.proto",
				),
				NewProtoFile(
					filepath.Join(setting_test.TestDataPath("testdir"), "/test.proto"),
					"../../../_testdata/testdir/test.proto",
				),
				NewProtoFile(
					filepath.Join(setting_test.TestDataPath("testdir"), "/test2.proto"),
					"../../../_testdata/testdir/test2.proto",
				),
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := NewProtoSet(test.inputTargetPaths, config.ExternalConfig{})
			if test.wantExistErr {
				if err == nil {
					t.Errorf("got err nil, but want err")
				}
				return
			}
			if err != nil {
				t.Errorf("got err %v, but want nil", err)
				return
			}

			for i, gotf := range got.ProtoFiles() {
				wantf := test.wantProtoFiles[i]
				if gotf.Path() != wantf.Path() {
					t.Errorf("got %v, but want %v", gotf.Path(), wantf.Path())
				}
				if gotf.DisplayPath() != wantf.DisplayPath() {
					t.Errorf("got %v, but want %v", gotf.DisplayPath(), wantf.DisplayPath())
				}
			}
		})
	}
}

func TestCollectAllProtoFiles_Excludes(t *testing.T) {
	externalConfig := config.ExternalConfig{
		Lint: config.Lint{
			Directories: config.Directories{
				Exclude: []string{
					"aa/dir1",
					"bb/dir2",
					"cc/dir1/dir2",
				},
				ExcludePattern: []string{
					"aaa/dir1",
					"/bbb/dir1",
					"ccc/dir1/**/",
					"ddd/dir1/**",
					"eee/**/dir2",
					"fff/d*r1",
					"ggg/d?r1",
					"hhh/[cd]ir1",
					"iii/{abc,dir,def}1",
				},
			},
			Files: config.Files{
				Exclude: []string{
					"path/to/file.proto",
					"/path/to/file2.proto",
					`path\to\file_windows.proto`,
				},
				ExcludePattern: []string{
					"an/other/dir31/*.proto",
				},
			},
		},
	}

	mockDirectoryWalker := func(root string, visit fs.WalkDirFunc) (err error) {
		paths := [][]string{
			{root, "dir1"},
			{root, "dir1", "dir2"},
			{root, "dir1", "dir2", "file1.proto"},
			{root, "dir1", "dir2", "file2.txt"},
			{root, "dir1", "dir2", "dir3"},
			{root, "dir1", "dir2", "dir3", "file1.proto"},
			{root, "dir1", "dir2", "dir3", "file3.proto"},
			{root, "dir1", "dir4"},
			{root, "dir1", "dir4", "file1.proto"},
			{root, "dir1", "dir4", "file4.proto"},
			{root, "dir2"},
			{root, "dir2", "file1.proto"},
		}
		skipped := make(map[string]bool)
		for _, p := range paths {
			joinedPath := strings.ReplaceAll(path.Join(p...), "/", string(os.PathSeparator))

			// TODO: improve skipping dirs (by using a tree structure?)
			for skippedPath := range skipped {
				if strings.HasPrefix(joinedPath, skippedPath) {
					continue
				}
			}

			var di fs.DirEntry
			err := visit(joinedPath, di, nil)
			if errors.Is(err, fs.SkipDir) {
				skipped[joinedPath] = true
			} else if err != nil {
				return err
			}
		}

		return err
	}

	tests := []struct {
		name           string
		absWorkDirPath string
		absPath        string
		expected       []string
	}{
		{
			name:           "",
			absWorkDirPath: "a/b",
			absPath:        "a/b/c/d",
			expected:       []string{"dir1/dir2/file1.proto", "dir1/dir2/dir3/file1.proto", "dir1/dir2/dir3/file3.proto", "dir1/dir4/file1.proto", "dir1/dir4/file4.proto", "dir2/file1.proto"},
		},
		{
			name:           "",
			absWorkDirPath: "a",
			absPath:        "a/aa",
			expected:       []string{"dir2/file1.proto"},
		},
		{
			name:           "",
			absWorkDirPath: "a",
			absPath:        "a/bb",
			expected:       []string{"dir1/dir2/file1.proto", "dir1/dir2/dir3/file1.proto", "dir1/dir2/dir3/file3.proto", "dir1/dir4/file1.proto", "dir1/dir4/file4.proto"},
		},
		{
			name:           "",
			absWorkDirPath: "a",
			absPath:        "a/cc",
			expected:       []string{"dir1/dir4/file1.proto", "dir1/dir4/file4.proto", "dir2/file1.proto"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			absWorkDirPath := strings.ReplaceAll(test.absWorkDirPath, "/", string(os.PathSeparator))
			absPath := strings.ReplaceAll(test.absPath, "/", string(os.PathSeparator))
			got, err := collectAllProtoFiles(absWorkDirPath, absPath, externalConfig, mockDirectoryWalker)
			if err != nil {
				t.Error("Did not expect an error")
			}

			if len(got) != len(test.expected) {
				t.Errorf("got %v, but want %v", len(got), len(test.expected))
			}

			for i, gotf := range got {
				wantf := NewProtoFile(
					strings.ReplaceAll(filepath.Join(absPath, test.expected[i]), "/", string(os.PathSeparator)),
					strings.ReplaceAll(filepath.Join(strings.Replace(absPath, absWorkDirPath+string(os.PathSeparator), "", 1), test.expected[i]), "/", string(os.PathSeparator)),
				)
				if gotf.Path() != wantf.Path() {
					t.Errorf("got %v, but want %v", gotf.Path(), wantf.Path())
				}
				if gotf.DisplayPath() != wantf.DisplayPath() {
					t.Errorf("got %v, but want %v", gotf.DisplayPath(), wantf.DisplayPath())
				}
			}
		})
	}
}
