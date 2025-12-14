package tos

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// CredentialsProvider fetches AK/SK/token with an expected expiration window.
// expires is the expected validity in seconds for the returned credentials.
type CredentialsProvider interface {
	Credentials(expires int64) (Credential, error)
}

// providerBackedCredentials wraps a CredentialsProvider to provide cached
// credentials with periodic asynchronous refresh and fallback to cached
// credentials when refresh fails.
type providerBackedCredentials struct {
	provider        CredentialsProvider
	expiresSeconds  int64
	refreshInterval time.Duration
	mu              sync.RWMutex
	cached          *providerCache
	stopCh          chan struct{}
}

type providerCache struct {
	cred      Credential
	expiredAt time.Time
	immortal  bool
}

func newProviderBackedCredentials(p CredentialsProvider, expiresSeconds int64, refreshInterval time.Duration) *providerBackedCredentials {
	w := &providerBackedCredentials{
		provider:        p,
		expiresSeconds:  expiresSeconds,
		refreshInterval: refreshInterval,
		stopCh:          make(chan struct{}),
	}
	// initial refresh attempt; failure tolerated (lazy fetch on use)
	_ = w.refresh()
	go w.loop()
	return w
}

func (w *providerBackedCredentials) loop() {
	ticker := time.NewTicker(w.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = w.refresh()
		case <-w.stopCh:
			return
		}
	}
}

// refresh tries to fetch new credentials; on failure, it marks cached creds immortal.
func (w *providerBackedCredentials) refresh() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	cred, err := w.provider.Credentials(w.expiresSeconds)
	if err != nil {
		if w.cached != nil {
			w.cached.immortal = true
		}
		return err
	}
	// cache validity is expiresSeconds minus 600s prefetch window
	validFor := time.Duration(w.expiresSeconds)*time.Second - 600*time.Second
	if validFor <= 0 {
		validFor = time.Duration(w.expiresSeconds) * time.Second
	}
	w.cached = &providerCache{
		cred:      cred,
		expiredAt: time.Now().Add(validFor),
		immortal:  false,
	}
	return nil
}

// Credential implements Credentials interface
func (w *providerBackedCredentials) Credential() Credential {
	w.mu.RLock()
	c := w.cached
	valid := c != nil && (c.immortal || time.Now().Before(c.expiredAt))
	if valid {
		cred := c.cred
		w.mu.RUnlock()
		return cred
	}
	w.mu.RUnlock()
	// try refresh synchronously (refresh has its own locking)
	_ = w.refresh()
	w.mu.RLock()
	c = w.cached
	var cred Credential
	if c != nil {
		cred = c.cred
	}
	w.mu.RUnlock()
	return cred
}

// Stop terminates background refresh goroutine.
func (w *providerBackedCredentials) Stop() {
	select {
	case <-w.stopCh:
		// already closed
		return
	default:
		close(w.stopCh)
	}
}

// WithCredentialsProvider sets a provider; the SDK wraps it with a periodic
// refresh wrapper that caches long-lived credentials and falls back to cache
// when refresh fails.
func WithCredentialsProvider(provider CredentialsProvider) ClientOption {
	return func(client *Client) {
		// Default long validity window: 10 hours
		expires := int64(10 * 60 * 60)
		// Default refresh interval: 10 minutes
		interval := 10 * time.Minute
		client.credentials = newProviderBackedCredentials(provider, expires, interval)
	}
}

// NOTE: Federation-based adapter removed; providerBackedCredentials implements
// periodic refresh and caching directly.

// StaticCredentialsProvider returns a fixed credential.
type StaticCredentialsProvider struct {
	cred Credential
}

func (p *StaticCredentialsProvider) Credentials(expires int64) (Credential, error) {
	return p.cred, nil
}

func NewStaticCredentialsProvider(ak, sk, securityToken string) CredentialsProvider {
	return &StaticCredentialsProvider{cred: Credential{AccessKeyID: ak, AccessKeySecret: sk, SecurityToken: securityToken}}
}

// EnvCredentialsProvider loads credentials from environment variables.
// TOS_ACCESS_KEY, TOS_SECRET_KEY, TOS_SECURITY_TOKEN(optional)
type EnvCredentialsProvider struct{}

func (p *EnvCredentialsProvider) Credentials(expires int64) (Credential, error) {
	ak := strings.TrimSpace(os.Getenv("TOS_ACCESS_KEY"))
	sk := strings.TrimSpace(os.Getenv("TOS_SECRET_KEY"))
	token := strings.TrimSpace(os.Getenv("TOS_SECURITY_TOKEN"))
	if ak == "" || sk == "" {
		return Credential{}, errors.New("tos: env credentials not found: require TOS_ACCESS_KEY and TOS_SECRET_KEY")
	}
	return Credential{AccessKeyID: ak, AccessKeySecret: sk, SecurityToken: token}, nil
}

// EcsCredentialsProvider fetches temporary credentials from Volcengine ECS MetaService.
// Default URL: http://100.96.0.96/volcstack/latest/iam/security_credentials/{role_name}
type EcsCredentialsProvider struct {
	roleName   string
	url        string
	httpClient *http.Client
}

// NewEcsCredentialsProvider creates provider with required roleName and optional custom URL.
// If url is empty, the default MetaService URL is used.
func NewEcsCredentialsProvider(roleName string, urlOpt ...string) CredentialsProvider {
	u := ""
	if len(urlOpt) > 0 {
		u = urlOpt[0]
	}
	if u == "" {
		u = "http://100.96.0.96/volcstack/latest/iam/security_credentials/{role_name}"
	}
	return &EcsCredentialsProvider{
		roleName:   roleName,
		url:        u,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *EcsCredentialsProvider) Credentials(expires int64) (Credential, error) {
	u := p.url
	if strings.Contains(u, "{role_name}") {
		u = strings.ReplaceAll(u, "{role_name}", p.roleName)
	} else if strings.HasSuffix(u, "/") {
		u = u + p.roleName
	}
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return Credential{}, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return Credential{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Credential{}, fmt.Errorf("tos: ecs meta service status %d", resp.StatusCode)
	}
	var out struct {
		AccessKeyId     string `json:"AccessKeyId"`
		SecretAccessKey string `json:"SecretAccessKey"`
		SessionToken    string `json:"SessionToken"`
		ExpiredTime     string `json:"ExpiredTime"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return Credential{}, err
	}
	if out.AccessKeyId == "" || out.SecretAccessKey == "" {
		return Credential{}, errors.New("tos: ecs provider returned empty ak/sk")
	}
	// ExpiredTime is informative; caching is handled by FederationCredentials using requested expires.
	return Credential{AccessKeyID: out.AccessKeyId, AccessKeySecret: out.SecretAccessKey, SecurityToken: out.SessionToken}, nil
}
