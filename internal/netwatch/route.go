package netwatch

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"
)

// Route is the address and country one service saw a request come from.
type Route struct {
	IP, Country string
}

// parseTrace reads a Cloudflare /cdn-cgi/trace answer ("ip=1.2.3.4\nloc=NL\n...").
func parseTrace(body string) (Route, error) {
	var r Route
	for _, line := range strings.Split(body, "\n") {
		k, v, ok := strings.Cut(strings.TrimSpace(line), "=")
		if !ok {
			continue
		}
		switch k {
		case "ip":
			a, err := netip.ParseAddr(v)
			if err != nil {
				return Route{}, err
			}
			r.IP = a.Unmap().String()
		case "loc":
			// XX and T1 (Tor) are not countries: leave the country empty.
			if len(v) == 2 && isLetters(v) && v != "XX" {
				r.Country = strings.ToUpper(v)
			}
		}
	}
	if r.IP == "" {
		return Route{}, errors.New("no ip= in the answer")
	}
	return r, nil
}

// Trace asks a service which address and country it sees. It goes over IPv4
// through the system proxy, the same way the protected apps reach the service.
func Trace(ctx context.Context, u string, timeout time.Duration) (Route, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	c := client(SystemProxy(), timeout, "tcp4")
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return Route{}, err
	}
	req.Header.Set("User-Agent", "curl/8.7.1")
	resp, err := c.Do(req)
	if err != nil {
		return Route{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return Route{}, errors.New(resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return Route{}, err
	}
	return parseTrace(string(b))
}

// Routes checks all the URLs at once; the ones that failed are left out.
func Routes(ctx context.Context, urls []string, timeout time.Duration) map[string]Route {
	out := map[string]Route{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, u := range urls {
		wg.Add(1)
		go func(u string) {
			defer wg.Done()
			if r, err := Trace(ctx, u, timeout); err == nil {
				mu.Lock()
				out[u] = r
				mu.Unlock()
			}
		}(u)
	}
	wg.Wait()
	return out
}
