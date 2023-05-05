// SPDX-FileCopyrightText: 2023 Sascha Brawer <sascha@brawer.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFindEntitiesDump(t *testing.T) {
	dumpsDir := t.TempDir()
	dir := filepath.Join(dumpsDir, "wikidatawiki", "entities")
	if err := os.MkdirAll(filepath.Join(dir, "20250215"), 0755); err != nil {
		t.Error(err)
		return
	}

	dumpPath := filepath.Join(dir, "20250215", "wikidata-20250215-all.json.bz2")
	if f, err := os.Create(dumpPath); err == nil {
		f.Close()
	} else {
		t.Error(err)
		return
	}

	err := os.Symlink(filepath.Join("20250215", "wikidata-20250215-all.json.bz2"),
		filepath.Join(dir, "latest-all.json.bz2"))
	if err != nil {
		t.Error(err)
		return
	}

	wantPath := filepath.Join(dir, "20250215", "wikidata-20250215-all.json.bz2")
	date, path, err := findEntitiesDump(dumpsDir)
	if err != nil {
		t.Error(err)
		return
	}

	if d := date.Format("2006-01-02"); d != "2025-02-15" {
		t.Errorf("want 2025-02-15, got %s", d)
	}

	got, _ := os.Stat(path)
	want, _ := os.Stat(wantPath)
	if !os.SameFile(want, got) {
		t.Errorf("want %q, got %q", wantPath, path)
	}
}

func TestExtractQuickStatements(t *testing.T) {
	inpath := filepath.Join("testdata", "entities.bz2")
	var out bytes.Buffer
	if err := extractQuickStatements(inpath, &out); err != nil {
		t.Error(err)
		return
	}
}
