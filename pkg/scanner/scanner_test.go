package scanner

import "testing"

func TestParseLsofLine_PreservesCommandFallback(t *testing.T) {
	ps := NewProcessScanner()

	record, err := ps.parseLsofLine("node 12345 kirby 22u IPv4 0x1234567890 0t0 TCP *:5173 (LISTEN)")
	if err != nil {
		t.Fatalf("parseLsofLine returned error: %v", err)
	}
	if record == nil {
		t.Fatal("expected record")
	}
	if record.Command != "node" {
		t.Fatalf("expected command fallback %q, got %q", "node", record.Command)
	}
	if record.Port != 5173 {
		t.Fatalf("expected port 5173, got %d", record.Port)
	}
}
