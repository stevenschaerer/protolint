package config_test

import (
	"testing"

	"github.com/yoheimuta/protolint/internal/filepathutil"
	"github.com/yoheimuta/protolint/linter/autodisable"

	"github.com/yoheimuta/protolint/internal/cmd/subcmds"
	"github.com/yoheimuta/protolint/internal/linter/config"
)

func TestExternalConfig_ShouldSkipRule(t *testing.T) {
	noDefaultExternalConfig := config.ExternalConfig{
		Lint: config.Lint{
			Ignores: []config.Ignore{
				{
					ID: "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
					Files: []string{
						"path/to/foo.proto",
						"/path/to/bar.proto",
						`\path\to\bar_windows.proto`,
					},
				},
				{
					ID: "ENUM_NAMES_UPPER_CAMEL_CASE",
					Files: []string{
						"path/to/foo.proto",
					},
				},
			},
			Directories: config.Directories{
				Exclude: []string{
					"path/to/dir",
					"/path/to/dir2",
					`\path\to\dir_windows`,
				},
				ExcludePattern: []string{
					"some/other/dir4",
					"/some/other/dir5",
					"/some/other/dir_windows",
					"yet/another/**/",
					"**/dir6",
					"some/other/d*7",
					"some/**/d*8",
					"some/other/di?9",
					"some/other/[cd]ir10",
					"some/other/{abc,dir,def}11",
					`\some\other\dir_w\[indows`,
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
			Rules: struct {
				NoDefault  bool     `yaml:"no_default"`
				AllDefault bool     `yaml:"all_default"`
				Add        []string `yaml:"add"`
				Remove     []string `yaml:"remove"`
			}{
				NoDefault: true,
				Add: []string{
					"FIELD_NAMES_LOWER_SNAKE_CASE",
					"MESSAGE_NAMES_UPPER_CAMEL_CASE",
					"ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
					"ENUM_NAMES_UPPER_CAMEL_CASE",
				},
				Remove: []string{
					"RPC_NAMES_UPPER_CAMEL_CASE",
				},
			},
		},
	}

	defaultExternalConfig := config.ExternalConfig{
		Lint: config.Lint{
			Ignores: []config.Ignore{
				{
					ID: "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
					Files: []string{
						"path/to/foo.proto",
						"path/to/bar.proto",
					},
				},
				{
					ID: "ENUM_NAMES_UPPER_CAMEL_CASE",
					Files: []string{
						"path/to/foo.proto",
					},
				},
			},
			Rules: struct {
				NoDefault  bool     `yaml:"no_default"`
				AllDefault bool     `yaml:"all_default"`
				Add        []string `yaml:"add"`
				Remove     []string `yaml:"remove"`
			}{
				NoDefault: false,
				Add: []string{
					"FIELD_NAMES_LOWER_SNAKE_CASE",
					"MESSAGE_NAMES_UPPER_CAMEL_CASE",
				},
				Remove: []string{
					"RPC_NAMES_UPPER_CAMEL_CASE",
				},
			},
		},
	}

	allRules, err := subcmds.NewAllRules(config.RulesOption{}, false, autodisable.Noop, false, nil)
	if err != nil {
		t.Error(err)
		return
	}

	for _, test := range []struct {
		name                        string
		externalConfig              config.ExternalConfig
		inputRuleID                 string
		inputDisplayPath            string
		inputDefaultRuleIDs         []string
		inputIsWindowsPathSeparator bool
		wantSkipRule                bool
	}{
		{
			name:             "ignore ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			inputDisplayPath: "path/to/foo.proto",
			wantSkipRule:     true,
		},
		{
			name:             "ignore ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			inputDisplayPath: "/path/to/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "ignore ENUM_NAMES_UPPER_CAMEL_CASE",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "ENUM_NAMES_UPPER_CAMEL_CASE",
			inputDisplayPath: "path/to/foo.proto",
			wantSkipRule:     true,
		},
		{
			name:                        "ignore a windows path by referring to a windows path",
			externalConfig:              noDefaultExternalConfig,
			inputRuleID:                 "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			inputDisplayPath:            `\path\to\bar_windows.proto`,
			inputIsWindowsPathSeparator: true,
			wantSkipRule:                true,
		},
		{
			name:                        "ignore a windows path by referring to an unix path",
			externalConfig:              noDefaultExternalConfig,
			inputRuleID:                 "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			inputDisplayPath:            `path\to\foo.proto`,
			inputIsWindowsPathSeparator: true,
			wantSkipRule:                true,
		},
		{
			name:             "not ignore FIELD_NAMES_LOWER_SNAKE_CASE",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "/path/to/bar.proto",
		},
		{
			name:             "not ignore ENUM_FIELD_NAMES_UPPER_SNAKE_CASE because of a file mismatch",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			inputDisplayPath: "path/to/baz.proto",
		},
		{
			name:             "not ignore ENUM_NAMES_UPPER_CAMEL_CASE because of a file mismatch",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "ENUM_NAMES_UPPER_CAMEL_CASE",
			inputDisplayPath: "path/to/bar.proto",
		},
		{
			name:             "not ignore an unix path by referring to a windows path",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "ENUM_FIELD_NAMES_UPPER_SNAKE_CASE",
			inputDisplayPath: `/path/to/bar_windows.proto`,
		},
		{
			name:           "not skip Add rules",
			externalConfig: noDefaultExternalConfig,
			inputRuleID:    "FIELD_NAMES_LOWER_SNAKE_CASE",
		},
		{
			name:           "skip noAdd rules",
			externalConfig: noDefaultExternalConfig,
			inputRuleID:    "RPC_NAMES_UPPER_CAMEL_CASE",
			wantSkipRule:   true,
		},
		{
			name:           "skip Remove rule",
			externalConfig: defaultExternalConfig,
			inputRuleID:    "RPC_NAMES_UPPER_CAMEL_CASE",
			wantSkipRule:   true,
		},
		{
			name:           "not skip noRemove rule",
			externalConfig: defaultExternalConfig,
			inputRuleID:    "FIELD_NAMES_LOWER_SNAKE_CASE",
		},
		{
			name:           "not skip default rules",
			externalConfig: defaultExternalConfig,
			inputRuleID:    "HOGE_RULE",
			inputDefaultRuleIDs: []string{
				"HOGE_RULE",
			},
		},
		{
			name:                "not skip default one",
			externalConfig:      config.ExternalConfig{},
			inputRuleID:         allRules.Default().IDs()[0],
			inputDefaultRuleIDs: allRules.Default().IDs(),
		},
		{
			name:             "exclude the directory",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "path/to/dir/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "exclude the another directory",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "/path/to/dir2/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "exclude the child directory",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "/path/to/dir2/child/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:                        "exclude the matched windows directory",
			externalConfig:              noDefaultExternalConfig,
			inputRuleID:                 "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath:            `\path\to\dir_windows\foo.proto`,
			inputIsWindowsPathSeparator: true,
			wantSkipRule:                true,
		},
		{
			name:                        "exclude the matched windows directory by referring to an unix filepath",
			externalConfig:              noDefaultExternalConfig,
			inputRuleID:                 "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath:            `\path\to\dir2\child\bar.proto`,
			inputIsWindowsPathSeparator: true,
			wantSkipRule:                true,
		},
		{
			name:             "not exclude the another directory",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "path/to/dir3/bar.proto",
		},
		{
			name:             "not exclude the unix directory by referring to a windows directory",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: `/path/to/dir_windows/bar.proto`,
		},
		{
			name:             "exclude the matched file",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "path/to/file.proto",
			wantSkipRule:     true,
		},
		{
			name:             "exclude the matched file",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "/path/to/file2.proto",
			wantSkipRule:     true,
		},
		{
			name:                        "exclude the matched windows file",
			externalConfig:              noDefaultExternalConfig,
			inputRuleID:                 "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath:            `path\to\file_windows.proto`,
			inputIsWindowsPathSeparator: true,
			wantSkipRule:                true,
		},
		{
			name:                        "exclude the matched windows file by referring to an unix file",
			externalConfig:              noDefaultExternalConfig,
			inputRuleID:                 "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath:            `\path\to\file2.proto`,
			inputIsWindowsPathSeparator: true,
			wantSkipRule:                true,
		},
		{
			name:             "not exclude the unmatched file",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "/path/to/file3.proto",
		},
		{
			name:             "not exclude the unmatched file path",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "path/to1/file.proto",
		},
		{
			name:             "not exclude the unix file by referring to a windows path",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: `path/to/file_windows.proto`,
		},
		{
			name:             "pattern exclude the directory",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/dir4/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the another directory",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "/some/other/dir5/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the child directory",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "/some/other/dir6/child/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:                        "pattern exclude a windows directory",
			externalConfig:              noDefaultExternalConfig,
			inputRuleID:                 "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath:            `\some\other\dir5\child\bar.proto`,
			inputIsWindowsPathSeparator: true,
			wantSkipRule:                true,
		},
		{
			name:             "not exclude the another directory via pattern",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/dir3/bar.proto",
		},
		{
			name:             "pattern exclude the directory with ** and no subdir",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "yet/another/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with ** and multiple subdirs",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "yet/another/dir/sub/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with ** at beginning and no subdir",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "dir6/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with ** at beginning and multiple subdirs",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "a/b/c/dir6/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with *",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/dir7/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with * and **",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/dir8/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with ?",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/dir9/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with ?",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/djr9/bar.proto",
		},
		{
			name:             "pattern exclude the directory with char set",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/dir10/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with char set",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/cir10/bar.proto",
			wantSkipRule:     true,
		},
		{
			name:             "pattern exclude the directory with char class",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "some/other/bir10/bar.proto",
		},
		{
			name:             "pattern exclude a file with *",
			externalConfig:   noDefaultExternalConfig,
			inputRuleID:      "FIELD_NAMES_LOWER_SNAKE_CASE",
			inputDisplayPath: "an/other/dir31/bar.proto",
			wantSkipRule:     true,
		},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			osPathSep := '/'
			if test.inputIsWindowsPathSeparator {
				osPathSep = '\\'
			}
			prevOSPathSep := filepathutil.OSPathSeparator
			filepathutil.OSPathSeparator = osPathSep
			defer func() {
				filepathutil.OSPathSeparator = prevOSPathSep
			}()

			got := test.externalConfig.ShouldSkipRule(
				test.inputRuleID,
				test.inputDisplayPath,
				test.inputDefaultRuleIDs,
			)
			if got != test.wantSkipRule {
				t.Errorf("got %v, but want %v", got, test.wantSkipRule)
			}
		})
	}
}
