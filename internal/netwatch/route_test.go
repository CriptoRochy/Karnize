package netwatch

import "testing"

func TestParseTrace(t *testing.T) {
	r, err := parseTrace("fl=123\nh=claude.ai\nip=45.158.10.20\nts=1.2\nloc=de\ntls=TLSv1.3\n")
	if err != nil || r.IP != "45.158.10.20" || r.Country != "DE" {
		t.Fatalf("got %+v, %v", r, err)
	}
	r, err = parseTrace("ip=2a01:4f8::1\nloc=T1\n")
	if err != nil || r.IP != "2a01:4f8::1" || r.Country != "" {
		t.Fatalf("tor exit: got %+v, %v", r, err)
	}
	if _, err := parseTrace("<html>blocked</html>"); err == nil {
		t.Fatal("an answer without ip= must be an error")
	}
	if _, err := parseTrace("ip=not-an-ip\n"); err == nil {
		t.Fatal("a broken ip must be an error")
	}
}
