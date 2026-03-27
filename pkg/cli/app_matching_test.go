package cli

import "testing"

func TestCanMatchByPath(t *testing.T) {
	t.Run("matches unique shared root", func(t *testing.T) {
		if !canMatchByPath("/repo", "/repo", "/repo", "/repo", map[string]int{"/repo": 1}, map[string]int{"/repo": 1}) {
			t.Fatal("expected unique root/cwd match to be allowed")
		}
	})

	t.Run("rejects ambiguous shared root", func(t *testing.T) {
		if canMatchByPath("/repo", "/repo", "/repo", "/repo", map[string]int{"/repo": 2}, map[string]int{"/repo": 2}) {
			t.Fatal("expected ambiguous shared root/cwd match to be rejected")
		}
	})

	t.Run("rejects ambiguous root even when process matches", func(t *testing.T) {
		if canMatchByPath("/repo", "/repo", "/repo", "/other", map[string]int{"/repo": 2}, map[string]int{"/repo": 1}) {
			t.Fatal("expected ambiguous root match to be rejected")
		}
	})
}
