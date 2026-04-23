package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectSpecTasks(t *testing.T) {
	t.Parallel()

	specRoot := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(specRoot, "component-a.json"), []byte(`{}`), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(specRoot, "feature-a"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(specRoot, "feature-a", "spec-a.json"), []byte(`{}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specRoot, "ignored.yaml"), []byte("openapi: 3.0.0"), 0o644))

	targetDir := filepath.Join(specRoot, "out")
	tasks, err := collectSpecTasks(specRoot, targetDir)
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].targetDir < tasks[j].targetDir
	})
	require.Equal(t, filepath.Join(targetDir, "component-a"), tasks[0].targetDir)
	require.Equal(t, filepath.Join(targetDir, "feature-a", "spec-a"), tasks[1].targetDir)
}

func TestMoveCommonCandidateFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	specA := filepath.Join(root, "component-a")
	specB := filepath.Join(root, "feature-a", "spec-a")
	require.NoError(t, os.MkdirAll(specA, 0o755))
	require.NoError(t, os.MkdirAll(specB, 0o755))

	shared := []byte("package api\n\n// shared\n")
	uniqueA := []byte("package api\n\n// unique a\n")
	uniqueB := []byte("package api\n\n// unique b\n")

	require.NoError(t, os.WriteFile(filepath.Join(specA, "oas_cfg_gen.go"), shared, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specB, "oas_cfg_gen.go"), shared, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specA, "oas_schemas_gen.go"), uniqueA, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specB, "oas_schemas_gen.go"), uniqueB, 0o644))

	require.NoError(t, moveCommonCandidateFiles(root, []string{specA, specB}, commonGeneratedCandidates))

	_, err := os.Stat(filepath.Join(root, "common", "oas_cfg_gen.go"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(specA, "oas_cfg_gen.go"))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(specB, "oas_cfg_gen.go"))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(specA, "oas_schemas_gen.go"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(specB, "oas_schemas_gen.go"))
	require.NoError(t, err)
}

func TestParseCandidateSet(t *testing.T) {
	t.Parallel()

	got := parseCandidateSet(" oas_cfg_gen.go , oas_json_gen.go ,, ")
	require.Len(t, got, 2)
	_, ok := got["oas_cfg_gen.go"]
	require.True(t, ok)
	_, ok = got["oas_json_gen.go"]
	require.True(t, ok)
}

func TestMoveCommonCandidateFiles_IgnoresExistingCommonDir(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	specA := filepath.Join(root, "component-a")
	specB := filepath.Join(root, "feature-a", "spec-a")
	commonDir := filepath.Join(root, "common")
	require.NoError(t, os.MkdirAll(specA, 0o755))
	require.NoError(t, os.MkdirAll(specB, 0o755))
	require.NoError(t, os.MkdirAll(commonDir, 0o755))

	shared := []byte("package api\n\n// shared cfg\n")
	require.NoError(t, os.WriteFile(filepath.Join(specA, "oas_cfg_gen.go"), shared, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specB, "oas_cfg_gen.go"), shared, 0o644))
	// Simulate previous run output.
	require.NoError(t, os.WriteFile(filepath.Join(commonDir, "oas_cfg_gen.go"), shared, 0o644))

	require.NoError(t, moveCommonCandidateFiles(root, []string{specA, specB}, map[string]struct{}{
		"oas_cfg_gen.go": {},
	}))

	_, err := os.Stat(filepath.Join(commonDir, "oas_cfg_gen.go"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(specA, "oas_cfg_gen.go"))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(specB, "oas_cfg_gen.go"))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))
}

func TestMoveCommonCandidateFiles_StaleOutputFolderDoesNotAffectCurrentRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	specA := filepath.Join(root, "component-a")
	specB := filepath.Join(root, "feature-a", "spec-a")
	stale := filepath.Join(root, "old-spec")
	require.NoError(t, os.MkdirAll(specA, 0o755))
	require.NoError(t, os.MkdirAll(specB, 0o755))
	require.NoError(t, os.MkdirAll(stale, 0o755))

	shared := []byte("package api\n\n// shared cfg\n")
	require.NoError(t, os.WriteFile(filepath.Join(specA, "oas_cfg_gen.go"), shared, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specB, "oas_cfg_gen.go"), shared, 0o644))
	// Simulate stale output from old run/spec that should be ignored for current task count.
	require.NoError(t, os.WriteFile(filepath.Join(stale, "oas_cfg_gen.go"), []byte("package api\n\n// stale\n"), 0o644))

	require.NoError(t, moveCommonCandidateFiles(root, []string{specA, specB}, map[string]struct{}{
		"oas_cfg_gen.go": {},
	}))

	_, err := os.Stat(filepath.Join(root, "common", "oas_cfg_gen.go"))
	require.NoError(t, err)
}

func TestMoveCommonCandidateFiles_UsesOverrideCandidateSet(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	specA := filepath.Join(root, "component-a")
	specB := filepath.Join(root, "feature-a", "spec-a")
	require.NoError(t, os.MkdirAll(specA, 0o755))
	require.NoError(t, os.MkdirAll(specB, 0o755))

	sharedCfg := []byte("package api\n\n// shared cfg\n")
	sharedJSON := []byte("package api\n\n// shared json\n")
	require.NoError(t, os.WriteFile(filepath.Join(specA, "oas_cfg_gen.go"), sharedCfg, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specB, "oas_cfg_gen.go"), sharedCfg, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specA, "oas_json_gen.go"), sharedJSON, 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(specB, "oas_json_gen.go"), sharedJSON, 0o644))

	require.NoError(t, moveCommonCandidateFiles(root, []string{specA, specB}, map[string]struct{}{
		"oas_json_gen.go": {},
	}))

	_, err := os.Stat(filepath.Join(root, "common", "oas_json_gen.go"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(root, "common", "oas_cfg_gen.go"))
	require.Error(t, err)
	require.True(t, os.IsNotExist(err))

	_, err = os.Stat(filepath.Join(specA, "oas_cfg_gen.go"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(specB, "oas_cfg_gen.go"))
	require.NoError(t, err)
}
