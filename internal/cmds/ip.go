// Package cmds provides the implementation backing token-forge's cli.
// This section of the cmds package holds the implementation for the login
// command.
package cmds

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// treat this like a const.
var ipCheckURLs = []string{"https://ifconfig.me/ip", "https://ipinfo.tw/ip", "https://myexternalip.com/raw", "https://ipecho.net/plain", "https://icanhazip.com"}

// IPCmd represents the ip check cli command.
type IPCmd struct {
	Globals
	ProxyConfig
}

type ipResult struct {
	ips map[string]string
	mu  sync.Mutex
}

// Run the ip check based on the parameters of the IpCmd.
func (p *IPCmd) Run() error {
	if err := setProxy(p.Proxy); err != nil {
		return err
	}

	//nolint:exhaustruct
	r := &ipResult{
		ips: make(map[string]string, 0),
	}
	wg := sync.WaitGroup{}
	log.Println("checking public ip...")
	for _, u := range ipCheckURLs {
		wg.Add(1)
		go func(u string) {
			ip, err := doIPCheck(u)
			if err != nil {
				log.Printf("error: %v; continuing...", err)
			}
			r.mu.Lock()
			r.ips[u] = ip
			r.mu.Unlock()
			wg.Done()
		}(u)
	}
	wg.Wait()

	for k, v := range r.ips {
		log.Printf("--- '%s' reports ip as: %s", k, v)
	}

	return nil
}

func doIPCheck(u string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, u, nil)
	if err != nil {
		return "", fmt.Errorf("error building ip check request: %w", err)
	}

	rsp, err := http.DefaultClient.Do(req)
	if rsp != nil {
		defer func() {
			if herr := rsp.Body.Close(); herr != nil {
				err = fmt.Errorf("error closing response body: %w", herr)
			}
		}()
	}
	if err != nil {
		return "", fmt.Errorf("error checking public ip: %w", err)
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	return string(body), nil
}
