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
	"errors"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/EngFlow/gazelle_cc/internal/index"
	"github.com/EngFlow/gazelle_cc/language/cc"
	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/label"
	"github.com/bazelbuild/bazel-gazelle/resolve"
	"github.com/bazelbuild/bazel-gazelle/walk"
)

const usage = `Usage: local_index [flags] <workspace_path>

Walks the Bazel workspace at workspace_path, parses BUILD files, and writes a
DependencyIndex JSON file for use with the gazelle:cc_indexfile directive.
By default the index is written to output.json in the current working directory.

Flags:
`

func main() {
	workspacePath, outputPath, repoName := parseArgs()
	workspaceAbs := resolveWorkspace(workspacePath)
	depIndex, err := buildDependencyIndex(workspaceAbs, repoName)
	if err != nil {
		log.Fatalf("walk: %v", err)
	}
	if err := writeIndex(outputPath, depIndex); err != nil {
		log.Fatalf("write output: %v", err)
	}
}

func parseArgs() (workspacePath, outputPath, repoName string) {
	outputFlag := flag.String("output", "output.json", "path to write the DependencyIndex JSON file")
	repoNameFlag := flag.String("repo_name", "", "repository name for generated labels")
	flag.Usage = func() {
		os.Stderr.WriteString(usage)
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		log.Fatalf("expected 1 positional argument (workspace_path), got %d", len(args))
	}

	workspacePath = args[0]
	outputPath = *outputFlag
	repoName = *repoNameFlag
	return
}

func resolveWorkspace(workspacePath string) string {
	workspaceAbs, err := filepath.Abs(workspacePath)
	if err != nil {
		log.Fatalf("workspace path: %v", err)
	}
	if err := validateWorkspace(workspaceAbs); err != nil {
		log.Fatalf("workspace: %v", err)
	}
	return workspaceAbs
}

func buildDependencyIndex(workspaceAbs, repoName string) (index.DependencyIndex, error) {
	cfg := newConfig(workspaceAbs, repoName)
	cexts, ccLang, err := setupWalkAndCheckFlags(cfg)
	if err != nil {
		return nil, err
	}

	depIndex := make(index.DependencyIndex)
	wf := func(args walk.Walk2FuncArgs) walk.Walk2FuncResult {
		if args.File == nil {
			return walk.Walk2FuncResult{}
		}
		for _, r := range args.File.Rules {
			specs := ccLang.Imports(args.Config, r, args.File)
			for _, spec := range specs {
				lbl := label.Label{
					Repo: args.Config.RepoName,
					Pkg:  args.File.Pkg,
					Name: r.Name(),
				}
				depIndex[spec.Imp] = append(depIndex[spec.Imp], lbl)
			}
		}
		return walk.Walk2FuncResult{}
	}

	if err := walk.Walk2(cfg, cexts, []string{workspaceAbs}, walk.VisitAllUpdateSubdirsMode, wf); err != nil {
		return nil, err
	}
	return depIndex, nil
}

func setupWalkAndCheckFlags(cfg *config.Config) ([]config.Configurer, resolve.Resolver, error) {
	walkCfg := &walk.Configurer{}
	resolveCfg := &resolve.Configurer{}
	ccLang := cc.NewLanguage()
	cexts := []config.Configurer{walkCfg, resolveCfg, ccLang}
	fs := flag.NewFlagSet("local", flag.ContinueOnError)
	walkCfg.RegisterFlags(fs, "fix", cfg)
	resolveCfg.RegisterFlags(fs, "fix", cfg)
	if err := walkCfg.CheckFlags(fs, cfg); err != nil {
		return nil, nil, err
	}
	if err := resolveCfg.CheckFlags(fs, cfg); err != nil {
		return nil, nil, err
	}
	return cexts, ccLang, nil
}

func writeIndex(outputPath string, depIndex index.DependencyIndex) error {
	data, err := json.MarshalIndent(depIndex, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, data, 0644)
}

func validateWorkspace(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return os.ErrNotExist
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		name := e.Name()
		if name == "WORKSPACE" || name == "WORKSPACE.bazel" || name == "MODULE.bazel" {
			return nil
		}
	}
	return errors.New("workspace root must contain WORKSPACE, WORKSPACE.bazel, or MODULE.bazel")
}

func newConfig(repoRoot, repoName string) *config.Config {
	c := config.New()
	c.RepoRoot = repoRoot
	c.WorkDir = repoRoot
	c.RepoName = repoName
	return c
}
