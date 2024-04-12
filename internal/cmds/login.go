// Package cmds provides the implementation backing token-forge's cli.
// This section of the cmds package holds the implementation for the login
// command.
package cmds

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/go-github/v61/github"
	"github.com/pyqlsa/token-forge/internal/bar"
	"github.com/pyqlsa/token-forge/internal/fileutil"
	"github.com/pyqlsa/token-forge/internal/ghtoken"
	"golang.org/x/oauth2"
)

const (
	gheURLPrefix       = "https://"
	gheBaseURLSuffix   = "/api/v3/"
	gheUploadURLSuffix = "/api/uploads/"
)

// LoginCmd represents the login w/ token(s) cli command.
type LoginCmd struct {
	Globals
	TokenSourceArgs
	TokenParams
	ProxyConfig
	ForceCheck bool   `help:"Force a check of the logged in user so the rate limit is decremented."                     short:"c"`
	Host       string `help:"The GitHub Enterprise hostname to interact with; if not specified, github.com is assumed."`
}

// GhUserInfo holds GitHub user information along with the token used to get
// the information.
type GhUserInfo struct {
	Token *ghtoken.GhToken `json:"token"`
	Info  *github.User     `json:"info"`
}

// Run the login test based on the parameters of the LoginCmd.
func (l *LoginCmd) Run() error {
	if err := setProxy(l.Proxy); err != nil {
		return err
	}

	if l.BatchSize < 1 {
		return fmt.Errorf("must specify a batch size of 1 or greater")
	}

	switch {
	case l.NoAuth:
		return testLoginWithTokens(context.Background(), l.Host, nilTokenSource(l.NumTokens), l.BatchSize, l.ForceCheck, l.Debug)
	case l.Generated:
		if len(l.Prefix) > 0 && !ghtoken.IsValidPrefix(l.Prefix) {
			return fmt.Errorf("prefix '%s' is not a valid token prefix", l.Prefix)
		}

		return testLoginWithTokens(context.Background(), l.Host, generatedTokenSource(l.Prefix, l.NumTokens), l.BatchSize, l.ForceCheck, l.Debug)
	case len(l.File) > 0:
		source, err := fileTokenSource(l.File, l.NumTokens)
		if err != nil {
			return err
		}

		return testLoginWithTokens(context.Background(), l.Host, source, l.BatchSize, l.ForceCheck, l.Debug)
	default:
		token := ghtoken.ParseGhToken(l.Token)
		if len(token.FullToken) < 1 {
			return fmt.Errorf("error: token '%s' is malformed", l.Token)
		}
		//nolint:exhaustruct
		source := &sTokenSource{
			tokens: []*ghtoken.GhToken{token},
		}

		return testLoginWithTokens(context.Background(), l.Host, source, l.BatchSize, l.ForceCheck, l.Debug)
	}
}

type testResult struct {
	msg       string
	err       error // reserved for fatal errors (should probably kill context).
	collision bool
	rate      *github.Rate
}

type testBundle struct {
	result *testResult
	client *github.Client
	tok    *ghtoken.GhToken
}

// Test login with the provided tokens. First, the rate limit api is tested to
// determine if the token is valid, then if the token is determined to be
// valid, the information for the current user is queried.  In the future,
// different api endpoints should be queried based on the type of token being
// tested.
func testLoginWithTokens(ctx context.Context, host string, source tokenSource, batchSize int, forceCheck, debug bool) error {
	log.Printf("testing w/ %d tokens", source.remaining())
	tokensLeft := source.remaining()
	progress := bar.NewBar(source.remaining())
	wg := sync.WaitGroup{}
	bundles := make(chan *testBundle, batchSize)
	// kick off the initial batch of workers
	for !source.done() && batchSize > 0 {
		wg.Add(1)
		token, err := source.pop()
		if err != nil {
			return fmt.Errorf("error popping token: %w", err)
		}
		go asyncTestTokenViaRateLimit(ctx, &wg, bundles, host, token)
		batchSize--
	}
	gotem := make([]*testBundle, 0)
	for b := range bundles {
		// as we're gathering results and the source hasn't been drained, kick off
		// new workers; this makes sure that concurrent workers aren't entirely
		// unbounded
		if !source.done() {
			wg.Add(1)
			token, err := source.pop()
			if err != nil {
				return fmt.Errorf("error popping token: %w", err)
			}
			go asyncTestTokenViaRateLimit(ctx, &wg, bundles, host, token)
		}
		if debug {
			log.Printf("result for token '%+v' --- ", b.tok)
			log.Printf("--- msg: %s", b.result.msg)
			log.Printf("--- err: %v", b.result.err)
			log.Printf("--- collision: %t", b.result.collision)
			log.Printf("--- rates: %v", b.result.rate)
		}
		if b.result.collision {
			gotem = append(gotem, b)
		}
		// TODO: figure out a better place to do this
		if forceCheck {
			checkCurrentUser(ctx, b.client, b.tok)
		}
		if err := progress.Inc(); err != nil {
			log.Printf("error adding to the progressbar? %v", err)
		}
		tokensLeft--
		if tokensLeft < 1 {
			break
		}
	}
	close(bundles)
	wg.Wait()
	if err := progress.Finish(); err != nil {
		log.Printf("error finishing the progressbar? %v", err)
	}

	// TODO: do this async rather than at the end
	if len(gotem) > 0 {
		log.Println("checking colliders...")
	} else {
		log.Println("sad.... no colliders")
	}
	for _, b := range gotem {
		observeRateLimit(b.result.rate)
		checkCurrentUser(ctx, b.client, b.tok)
	}

	return nil
}

func asyncTestTokenViaRateLimit(ctx context.Context, wg *sync.WaitGroup, bundles chan *testBundle, host string, token *ghtoken.GhToken) {
	client, err := newGithubClient(ctx, host, token)
	if err == nil {
		result := queryRateLimit(ctx, client, token)
		bundles <- &testBundle{
			client: client,
			tok:    token,
			result: result,
		}
	} else {
		bundles <- &testBundle{
			client: nil,
			tok:    token,
			result: &testResult{
				msg:       "derp",
				err:       fmt.Errorf("error instantiating client: %w", err),
				collision: false,
				rate:      &github.Rate{}, //nolint:exhaustruct
			},
		}
	}
	wg.Done()
}

// Test credentials populated in the github client for validity via the rate limit api.
func queryRateLimit(ctx context.Context, client *github.Client, token *ghtoken.GhToken) *testResult {
	var (
		msg  string
		rate *github.Rate
	)
	passed := false
	rateLimit, rsp, err := client.RateLimits(ctx)
	switch {
	case err != nil && rsp != nil:
		// TODO: probably also need to check rate limit to detect collision somewhere.
		if rsp.StatusCode != http.StatusUnauthorized {
			// got an error other than bad credentials.
			msg = fmt.Sprintf("error from client while getting rate limit: %v", err)
		} else {
			msg = fmt.Sprintf("bad credentials, don't check user: %v", err)
		}
		rate = &rsp.Rate
	case rateLimit != nil:
		passed = true
		rate = rateLimit.Core
		if token != nil {
			msg = fmt.Sprintf("no error from client, our token might be good: %s", token.FullToken)
		} else {
			msg = "no error from client, but no token used"
		}
	case err != nil:
		msg = fmt.Sprintf("got an error from the client and nil response: %v", err)
	default:
		msg = "fatal error: all nil values encountered when checking rate limit; will not proceed"
	}

	// 60 is the magic number for max unauthenticated requests per hour.
	// there's probably a better way of doing this, i.e. check limit while
	// unauthenticated, save value, then check if returned limit is ever greater.
	if rate != nil && rate.Limit > 60 {
		passed = true
		msg = fmt.Sprintf("possible collision detected via max rate limit check w/ token: %s", token.FullToken)
	}

	return &testResult{
		msg:       msg,
		err:       nil,
		collision: passed,
		rate:      rate,
	}
}

func checkCurrentUser(ctx context.Context, client *github.Client, token *ghtoken.GhToken) {
	usr, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Printf("got an error from the client: %v", err)
	}

	info := GhUserInfo{
		Token: token,
		Info:  usr,
	}

	if err := fileutil.WriteJSON(os.Stdout, info); err != nil {
		log.Printf("error parsing user info: %v", err)
	}
}

// Observe a GitHub rate limit.
func observeRateLimit(rate *github.Rate) {
	if rate.Remaining == 0 {
		reset := rate.Reset
		limit := rate.Limit
		log.Printf("rate limit of %d requests reached, waiting until %s", limit, reset.Time.String())
		time.Sleep(time.Until(reset.Time))
	} else {
		buf := bytes.NewBufferString("current rate limit info:\n")
		if err := fileutil.WriteJSON(buf, rate); err != nil {
			log.Printf("failed parsing rate limit info: %v", err)
		} else {
			log.Println(buf.String())
		}
	}
}

// Builds a new authenticated client for the given host with the given token;
// if the host is empty, github.com is assumed; if the token is  nil, the
// returned client is unauthenticated.
func newGithubClient(ctx context.Context, host string, token *ghtoken.GhToken) (*github.Client, error) {
	var (
		httpClient *http.Client
		client     *github.Client
		err        error
	)

	if token != nil {
		// log.Printf("using token '%s'", token.FullToken)
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token.FullToken}, //nolint:exhaustruct
		)
		httpClient = oauth2.NewClient(ctx, ts)
	}

	if len(host) > 0 {
		baseURL := fmt.Sprintf("%s%s%s", gheURLPrefix, host, gheBaseURLSuffix)
		uploadURL := fmt.Sprintf("%s%s%s", gheURLPrefix, host, gheUploadURLSuffix)
		client, err = github.NewEnterpriseClient(baseURL, uploadURL, httpClient)
		if err != nil {
			return nil, fmt.Errorf("failed to establish enterprise client for host '%s': %w", host, err)
		}
	} else {
		client = github.NewClient(httpClient)
	}

	return client, nil
}
