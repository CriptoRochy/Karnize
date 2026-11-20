package guard

import (
	"testing"
	"time"

	"github.com/ef1m03/kordon/internal/config"
	"github.com/ef1m03/kordon/internal/procs"
)

const (
	vpnIP    = "45.158.10.20" // what Claude's own servers see (through the VPN)
	directIP = "128.204.1.2"  // what some IP services see (they go around the VPN)
)

func claudeRoutes(ip, cc string) map[string]Route {
	p, _ := procs.PresetByID("claude")
	m := map[string]Route{}
	for _, u := range p.Routes {
		m[u] = Route{IP: ip, Country: cc}
	}
	return m
}

// splitProbe is what split routing looks like: the IP services disagree.
func splitProbe(routes map[string]Route) ProbeResult {
	return ProbeResult{IP: vpnIP, Country: "DE", Votes: 3, Others: []string{directIP}, Routes: routes}
}

func TestSplitRoutingTrustsTheServiceRoute(t *testing.T) {
	h := newHarness(t, vpnIP)
	h.step(0)
	h.probeFull(splitProbe(claudeRoutes(vpnIP, "DE")))
	h.want(OK)
	if h.p.frozen[claude.PID] {
		t.Fatal("claude frozen although its own servers see the allowed IP")
	}
}

func TestServiceRouteNotAllowedFreezes(t *testing.T) {
	h := newHarness(t, vpnIP)
	h.step(0)
	h.probe(vpnIP, "DE")
	h.want(OK)
	// The IP services agree on the VPN, but Claude itself is reached directly.
	h.probeFull(ProbeResult{IP: vpnIP, Country: "DE", Votes: 2, Routes: claudeRoutes(directIP, "RU")})
	h.want(Suspect)
	if s := h.g.Snapshot(h.now); s.Reason != IPNotAllowed {
		t.Fatalf("reason = %v, want IPNotAllowed", s.Reason)
	}
	if !h.p.frozen[claude.PID] {
		t.Fatal("claude not frozen while its route goes around the VPN")
	}
}

func TestMissingRouteFallsBackToIPServices(t *testing.T) {
	h := newHarness(t, vpnIP)
	h.step(0)
	routes := claudeRoutes(vpnIP, "DE")
	for u := range routes {
		delete(routes, u) // one route didn't answer
		break
	}
	h.probeFull(splitProbe(routes))
	h.want(Suspect)
	if s := h.g.Snapshot(h.now); s.Reason != Unconfirmed {
		t.Fatalf("reason = %v, want Unconfirmed", s.Reason)
	}
}

func TestRouteCheckOffUsesIPServices(t *testing.T) {
	h := newHarness(t, vpnIP)
	h.g.Update(h.now, func(c *config.Config) { c.RouteCheck = false })
	h.step(0)
	h.probeFull(splitProbe(claudeRoutes(vpnIP, "DE")))
	h.want(Suspect)
}

func TestAppWithoutRouteKeepsTheOldCheck(t *testing.T) {
	h := newHarness(t, vpnIP)
	// Safari has no route of its own, and it shares the rules with Claude.
	h.g.Update(h.now, func(c *config.Config) {
		c.CustomApps = append(c.CustomApps, config.CustomApp{Name: "Safari", Match: []string{"Safari"}})
	})
	h.step(0)
	h.probeFull(splitProbe(claudeRoutes(vpnIP, "DE")))
	h.want(Suspect)
}

func TestFirstRunAllowsTheServiceRouteNotTheDirectIP(t *testing.T) {
	h := newHarness(t)
	h.step(0)
	h.probeFull(ProbeResult{IP: directIP, Country: "RU", Votes: 2, Others: []string{vpnIP}, Routes: claudeRoutes(vpnIP, "DE")})
	h.want(OK)
	if h.saved == nil || len(h.saved.Allow.IPs) != 1 || h.saved.Allow.IPs[0] != vpnIP {
		t.Fatalf("first run must allow the route the app takes, got %+v", h.saved.Allow)
	}
}

func TestEmptiedAllowListIsNotRefilled(t *testing.T) {
	h := newHarness(t, vpnIP) // not a first run
	h.g.Update(h.now, func(c *config.Config) { c.Allow = config.Allow{} })
	h.step(0)
	h.probe(directIP, "RU")
	h.want(Suspect)
	if h.had(EvAutoAllowed) {
		t.Fatal("the home IP was allowed by itself after the list was emptied")
	}
}

func TestFreezeWithNoAppRunningIsNotCounted(t *testing.T) {
	h := newHarness(t, vpnIP)
	h.p.procs = nil // nothing protected is running
	h.step(0)
	h.probe(vpnIP, "DE")
	h.probe(directIP, "RU")
	h.want(Suspect)
	h.step(25 * time.Second)
	h.want(Blocked)
	c := h.g.Snapshot(h.now).Counters
	if c.Freezes != 0 || c.Blocks != 0 {
		t.Fatalf("counted a freeze with no app running: %+v", c)
	}
}
