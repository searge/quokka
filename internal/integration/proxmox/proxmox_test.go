package proxmox

import "testing"

func TestParseResourceIDExtractsID(t *testing.T) {
	p := New("forge-ovh-cli")

	id := p.parseResourceID("ok\nID: 321\ndone")
	if id != "321" {
		t.Fatalf("expected resource id 321, got %q", id)
	}
}

func TestParseResourceIDReturnsEmptyWhenMissing(t *testing.T) {
	p := New("forge-ovh-cli")

	id := p.parseResourceID("no id here")
	if id != "" {
		t.Fatalf("expected empty id, got %q", id)
	}
}
