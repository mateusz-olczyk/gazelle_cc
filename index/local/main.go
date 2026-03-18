// Copyright 2026 EngFlow Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/EngFlow/gazelle_cc/internal/index"
	"github.com/EngFlow/gazelle_cc/language/cc"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/language"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/walk"
)

const (
	cmd         = "local"
	gazelle_cmd = "fix"
	usage       = `Usage: ` + cmd + ` [flags]

Walks the Bazel workspace (current directory and repo root follow gazelle
flags, e.g. -repo_root), parses BUILD files, and writes a DependencyIndex JSON
file for use with the gazelle:cc_indexfile directive. By default the index is
written to output.json in the current working directory.

With -append, an existing output file is read first and the new index is
merged into it (union of header paths and labels).

Flags:
`
)

func main() {
	outputPath, appendMode, cfg, cexts, ccLang := parseArgs()

	depIndex, err := buildDependencyIndex(cfg, cexts, ccLang)
	if err != nil {
		log.Fatalf("walk: %v", err)
	}
	if err := writeIndex(outputPath, depIndex, appendMode); err != nil {
		log.Fatalf("write output: %v", err)
	}
}

func parseArgs() (
	outputPath string,
	appendMode bool,
	cfg *config.Config,
	cexts []config.Configurer,
	ccLang language.Language,
) {
	cfg = config.New()
	commonCfg := &config.CommonConfigurer{}
	walkCfg := &walk.Configurer{}
	resolveCfg := &resolve.Configurer{}
	ccLang = cc.NewLanguage()
	cexts = []config.Configurer{commonCfg, walkCfg, resolveCfg, ccLang}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cfg.WorkDir = wd

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	fs.SetOutput(os.Stderr)
	outputFlag := fs.String("output", "output.json", "path to write the DependencyIndex JSON file")
	appendFlag := fs.Bool("append", false, "merge into existing output file if present")
	repoNameFlag := fs.String("repo_name", "", "optional custom repository name for generated labels")
	commonCfg.RegisterFlags(fs, gazelle_cmd, cfg)
	walkCfg.RegisterFlags(fs, gazelle_cmd, cfg)
	resolveCfg.RegisterFlags(fs, gazelle_cmd, cfg)
	fs.Usage = func() {
		os.Stderr.WriteString(usage)
		fs.PrintDefaults()
	}

	fs.Parse(os.Args[1:])
	if args := fs.Args(); len(args) > 0 {
		log.Fatalf("unexpected positional argument(s): %q", args)
	}
	if err := commonCfg.CheckFlags(fs, cfg); err != nil {
		log.Fatalf("flags: %v", err)
	}
	if err := walkCfg.CheckFlags(fs, cfg); err != nil {
		log.Fatalf("flags: %v", err)
	}
	if err := resolveCfg.CheckFlags(fs, cfg); err != nil {
		log.Fatalf("flags: %v", err)
	}
	if *repoNameFlag != "" {
		cfg.RepoName = *repoNameFlag
	}

	return *outputFlag, *appendFlag, cfg, cexts, ccLang
}

func buildDependencyIndex(
	cfg *config.Config,
	cexts []config.Configurer,
	ccLang language.Language,
) (index.DependencyIndex, error) {
	var depIndex index.DependencyIndex
	wf := func(args walk.Walk2FuncArgs) walk.Walk2FuncResult {
		depIndex.Merge(indexBazelPackage(args, ccLang))
		return walk.Walk2FuncResult{}
	}

	if err := walk.Walk2(cfg, cexts, []string{cfg.RepoRoot}, walk.VisitAllUpdateSubdirsMode, wf); err != nil {
		return nil, err
	}
	return depIndex, nil
}

func indexBazelPackage(args walk.Walk2FuncArgs, lang language.Language) index.DependencyIndex {
	if args.File == nil {
		return nil
	}
	out := make(index.DependencyIndex)
	for _, r := range args.File.Rules {
		specs := lang.Imports(args.Config, r, args.File)
		for _, spec := range specs {
			lbl := label.Label{
				Repo: args.Config.RepoName,
				Pkg:  args.File.Pkg,
				Name: r.Name(),
			}
			out[spec.Imp] = append(out[spec.Imp], lbl)
		}
	}
	return out
}

func writeIndex(outputPath string, depIndex index.DependencyIndex, appendMode bool) error {
	out := depIndex
	if appendMode {
		data, err := os.ReadFile(outputPath)
		switch {
		case err == nil:
			var existing index.DependencyIndex
			if err := json.Unmarshal(data, &existing); err != nil {
				return fmt.Errorf("parse existing index %q: %w", outputPath, err)
			}
			existing.Merge(depIndex)
			out = existing
		case os.IsNotExist(err):
			// nothing to merge
		default:
			return err
		}
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}
