/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package translator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestNoReservedKeyAmbiguity scans every i18n YAML file and fails if any
// container node uses a go-i18n reserved key ("id", "description", "hash",
// "other", etc.) with a string value while also having map children.
// That combination makes go-i18n treat the container as a leaf message,
// which silently breaks all sibling sub-keys.
func TestNoReservedKeyAmbiguity(t *testing.T) {
	bundleDir := filepath.Join("..", "..", "..", "i18n")
	entries, err := os.ReadDir(bundleDir)
	if err != nil {
		t.Fatalf("cannot read i18n directory: %s", err)
	}

	for _, f := range entries {
		if f.IsDir() || filepath.Ext(f.Name()) != ".yaml" || f.Name() == "i18n.yaml" {
			continue
		}
		buf, err := os.ReadFile(filepath.Join(bundleDir, f.Name()))
		if err != nil {
			t.Fatalf("read %s: %s", f.Name(), err)
		}
		var root map[string]any
		if err := yaml.Unmarshal(buf, &root); err != nil {
			t.Fatalf("parse %s: %s", f.Name(), err)
		}
		assertNoReservedKeyAmbiguity(t, root, f.Name(), nil)
	}
}

func assertNoReservedKeyAmbiguity(t *testing.T, node map[string]any, file string, path []string) {
	t.Helper()
	reserved := map[string]bool{
		"id": true, "description": true, "hash": true,
		"leftdelim": true, "rightdelim": true,
		"zero": true, "one": true, "two": true,
		"few": true, "many": true, "other": true,
	}

	var reservedStringKeys []string
	hasMapChild := false
	for k, v := range node {
		if reserved[k] {
			if _, ok := v.(string); ok {
				reservedStringKeys = append(reservedStringKeys, k)
			}
		}
		if _, ok := v.(map[string]any); ok {
			hasMapChild = true
		}
	}
	if len(reservedStringKeys) > 0 && hasMapChild {
		t.Errorf("%s: node at path %q has reserved string key(s) %v alongside map children — "+
			"go-i18n will misinterpret this node as a message instead of a container",
			file, formatPath(path), reservedStringKeys)
	}

	// Recurse into map children.
	for k, v := range node {
		if sub, ok := v.(map[string]any); ok {
			assertNoReservedKeyAmbiguity(t, sub, file, append(append([]string{}, path...), k))
		}
	}
}

func formatPath(path []string) string {
	if len(path) == 0 {
		return "(root)"
	}
	return fmt.Sprintf("%s", strings.Join(path, "."))
}

// TestNoReservedKeyAmbiguity_CatchesWebhooksBug reproduces the exact pattern
// that broke the webhooks translation: a container node with "description"
// (a go-i18n reserved key) as a direct string child alongside map children.
// go-i18n's isMessage() treats the whole node as a leaf message, so the map
// children (e.g. "form") become invalid and parsing fails silently.
func TestNoReservedKeyAmbiguity_CatchesWebhooksBug(t *testing.T) {
	// This is the structure that caused the bug:
	//   webhooks:
	//     title: Webhooks
	//     description: Webhooks allow...   ← reserved key with string value
	//     form:                             ← map child → ambiguity!
	//       name: Name
	buggy := map[string]any{
		"ui": map[string]any{
			"admin": map[string]any{
				"webhooks": map[string]any{
					"title":       "Webhooks",
					"description": "Webhooks allow external services...",
					"form": map[string]any{
						"name": "Name",
					},
				},
			},
		},
	}

	errs := collectReservedKeyAmbiguities(buggy, "buggy.yaml", nil)
	if len(errs) == 0 {
		t.Fatal("expected ambiguity to be detected in the buggy structure, but got none")
	}
	t.Logf("correctly detected %d ambiguity(ies): %v", len(errs), errs)

	// Verify that renaming "description" → "desc" fixes the ambiguity.
	fixed := map[string]any{
		"ui": map[string]any{
			"admin": map[string]any{
				"webhooks": map[string]any{
					"title": "Webhooks",
					"desc":  "Webhooks allow external services...",
					"form": map[string]any{
						"name": "Name",
					},
				},
			},
		},
	}
	errs = collectReservedKeyAmbiguities(fixed, "fixed.yaml", nil)
	if len(errs) > 0 {
		t.Fatalf("expected no ambiguity after fix, but got: %v", errs)
	}
}

// collectReservedKeyAmbiguities walks the YAML tree and returns a list of
// ambiguity descriptions (same logic as assertNoReservedKeyAmbiguity but
// without requiring a testing.T, so it can be used to *expect* failures).
func collectReservedKeyAmbiguities(node map[string]any, file string, path []string) []string {
	reserved := map[string]bool{
		"id": true, "description": true, "hash": true,
		"leftdelim": true, "rightdelim": true,
		"zero": true, "one": true, "two": true,
		"few": true, "many": true, "other": true,
	}

	var errs []string
	var reservedStringKeys []string
	hasMapChild := false
	for k, v := range node {
		if reserved[k] {
			if _, ok := v.(string); ok {
				reservedStringKeys = append(reservedStringKeys, k)
			}
		}
		if _, ok := v.(map[string]any); ok {
			hasMapChild = true
		}
	}
	if len(reservedStringKeys) > 0 && hasMapChild {
		errs = append(errs, fmt.Sprintf("%s: node at path %q has reserved string key(s) %v alongside map children",
			file, formatPath(path), reservedStringKeys))
	}
	for k, v := range node {
		if sub, ok := v.(map[string]any); ok {
			errs = append(errs, collectReservedKeyAmbiguities(sub, file, append(append([]string{}, path...), k))...)
		}
	}
	return errs
}

// TestNewTranslator_MissingDefaultLanguage verifies that NewTranslator returns
// a clear error when the default language file (en_US.yaml) is absent from the
// bundle directory. This is the root-cause defense against the nil-Localizer
// panic in pacman's TrWithData.
func TestNewTranslator_MissingDefaultLanguage(t *testing.T) {
	// Create a temp bundle dir with only i18n.yaml and a non-default language.
	dir := t.TempDir()

	// Write the i18n metadata file.
	i18nContent := `language_options:
  - label: "French"
    value: "fr_FR"
    progress: 100
`
	if err := os.WriteFile(filepath.Join(dir, "i18n.yaml"), []byte(i18nContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a valid non-default language file (fr_FR) but NO en_US.yaml.
	frContent := `backend:
  base:
    success:
      other: "Succès."
`
	if err := os.WriteFile(filepath.Join(dir, "fr_FR.yaml"), []byte(frContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := NewTranslator(&I18n{BundleDir: dir})
	if err == nil {
		t.Fatal("expected error when default language file is missing, got nil")
	}
	if !strings.Contains(err.Error(), "default language") {
		t.Fatalf("error should mention default language, got: %s", err)
	}
}

// TestNewTranslator_WithDefaultLanguage verifies that NewTranslator succeeds
// when a valid en_US.yaml is present.
func TestNewTranslator_WithDefaultLanguage(t *testing.T) {
	dir := t.TempDir()

	i18nContent := `language_options:
  - label: "English"
    value: "en_US"
    progress: 100
`
	if err := os.WriteFile(filepath.Join(dir, "i18n.yaml"), []byte(i18nContent), 0644); err != nil {
		t.Fatal(err)
	}

	enContent := `backend:
  base:
    success:
      other: "Success."
    unknown:
      other: "Unknown error."
`
	if err := os.WriteFile(filepath.Join(dir, "en_US.yaml"), []byte(enContent), 0644); err != nil {
		t.Fatal(err)
	}

	tr, err := NewTranslator(&I18n{BundleDir: dir})
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil translator")
	}
}

// TestNewTranslator_I18nYamlNotPassedToAddTranslator verifies that i18n.yaml
// (the metadata file) is excluded from translation loading and does not cause
// an error or pollute the translator.
func TestNewTranslator_I18nYamlNotPassedToAddTranslator(t *testing.T) {
	dir := t.TempDir()

	// i18n.yaml has no backend section — it must be skipped.
	i18nContent := `language_options:
  - label: "English"
    value: "en_US"
    progress: 100
`
	if err := os.WriteFile(filepath.Join(dir, "i18n.yaml"), []byte(i18nContent), 0644); err != nil {
		t.Fatal(err)
	}

	enContent := `backend:
  base:
    success:
      other: "Success."
`
	if err := os.WriteFile(filepath.Join(dir, "en_US.yaml"), []byte(enContent), 0644); err != nil {
		t.Fatal(err)
	}

	tr, err := NewTranslator(&I18n{BundleDir: dir})
	if err != nil {
		t.Fatalf("expected no error, got: %s", err)
	}
	if tr == nil {
		t.Fatal("expected non-nil translator")
	}
}
