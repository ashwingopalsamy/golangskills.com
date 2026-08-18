package corpus

import (
	"strings"
	"testing"
)

func TestParseFrontmatter(t *testing.T) {
	t.Parallel()

	content := []byte(`---
name: example-skill
description: "Use for a: b. Do not use for c."
license: Apache-2.0
compatibility: Go 1.24 or newer.
---

# Example
`)
	frontmatter, body, err := parseFrontmatter(content)
	if err != nil {
		t.Fatalf("parseFrontmatter() error = %v", err)
	}
	if frontmatter.Name != "example-skill" {
		t.Fatalf("Name = %q, want example-skill", frontmatter.Name)
	}
	if !strings.Contains(body, "# Example") {
		t.Fatalf("body = %q, want heading", body)
	}
}

func TestParseFrontmatterRejectsUnknownKey(t *testing.T) {
	t.Parallel()

	content := []byte(`---
name: example-skill
description: "Use for x. Do not use for y."
license: Apache-2.0
compatibility: Go 1.24 or newer.
allowed-tools: Bash
---
`)
	_, _, err := parseFrontmatter(content)
	if err == nil || !strings.Contains(err.Error(), "unsupported frontmatter key") {
		t.Fatalf("parseFrontmatter() error = %v, want unsupported-key error", err)
	}
}

func TestDecodeStrictRejectsUnknownField(t *testing.T) {
	t.Parallel()

	var target struct {
		Name string `json:"name"`
	}
	err := decodeStrict([]byte(`{"name":"ok","unexpected":true}`), &target)
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("decodeStrict() error = %v, want unknown-field error", err)
	}
}
