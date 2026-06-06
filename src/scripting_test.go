/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dop251/goja"
)

func TestLoadScriptsFromDirectoryOnlyExecutesJavascriptFiles(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	hidden := filepath.Join(root, ".hidden")

	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.Mkdir(hidden, 0o755); err != nil {
		t.Fatal(err)
	}

	files := map[string]string{
		filepath.Join(root, "root.js"):          "globalThis.loadedScripts = (globalThis.loadedScripts || 0) + 1;",
		filepath.Join(nested, "nested.js"):      "globalThis.loadedScripts = (globalThis.loadedScripts || 0) + 1;",
		filepath.Join(root, "README.md"):        "this is not valid javascript",
		filepath.Join(root, "backup.js~"):       "this is not valid javascript",
		filepath.Join(root, ".hidden.js"):       "this is not valid javascript",
		filepath.Join(hidden, "ignored.js"):     "this is not valid javascript",
		filepath.Join(nested, ".ignored.swp"):   "this is not valid javascript",
		filepath.Join(nested, "generated.json"): "this is not valid javascript",
	}

	for path, contents := range files {
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	game := &Game{vm: goja.New()}

	if err := game.LoadScriptsFromDirectory(root); err != nil {
		t.Fatal(err)
	}

	if got := game.vm.Get("loadedScripts").ToInteger(); got != 2 {
		t.Fatalf("loadedScripts = %d, want 2", got)
	}
}
