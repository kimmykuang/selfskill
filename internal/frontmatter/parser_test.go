package frontmatter

import (
	"testing"
)

func TestParse_WithFrontmatter(t *testing.T) {
	content := []byte(`---
name: video-summary
version: 1.0.0
description: 视频内容总结
---
# Video Summary

This is the body.
`)
	meta, body, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta["name"] != "video-summary" {
		t.Errorf("expected name=video-summary, got %v", meta["name"])
	}
	if meta["version"] != "1.0.0" {
		t.Errorf("expected version=1.0.0, got %v", meta["version"])
	}
	if body != "# Video Summary\n\nThis is the body.\n" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestParse_NoFrontmatter(t *testing.T) {
	content := []byte("# Just a heading\n\nSome content.\n")
	meta, body, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta != nil {
		t.Errorf("expected nil metadata, got %v", meta)
	}
	if body != "# Just a heading\n\nSome content.\n" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestParse_EmptyFrontmatter(t *testing.T) {
	content := []byte("---\n---\n# Body\n")
	meta, body, err := Parse(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(meta) != 0 {
		t.Errorf("expected empty metadata, got %v", meta)
	}
	if body != "# Body\n" {
		t.Errorf("unexpected body: %q", body)
	}
}

func TestMarshal_WithMetadata(t *testing.T) {
	meta := map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
	}
	body := "# Content\n"
	result, err := Marshal(meta, body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse it back
	parsedMeta, parsedBody, err := Parse(result)
	if err != nil {
		t.Fatalf("unexpected error parsing marshaled content: %v", err)
	}
	if parsedMeta["name"] != "test" {
		t.Errorf("roundtrip name mismatch: %v", parsedMeta["name"])
	}
	if parsedBody != body {
		t.Errorf("roundtrip body mismatch: %q vs %q", parsedBody, body)
	}
}

func TestMarshal_NilMetadata(t *testing.T) {
	result, err := Marshal(nil, "just body")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result) != "just body" {
		t.Errorf("expected plain body, got %q", string(result))
	}
}
