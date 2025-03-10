package selfupdate

import (
	"context"
	"net/http"
	"os"

	"github.com/google/go-github/github"
	gitconfig "github.com/tcnksm/go-gitconfig"
	"golang.org/x/oauth2"
	"crypto/tls"
)

// Updater is responsible for managing the context of self-update.
// It contains GitHub client and its context.
type Updater struct {
	api    *github.Client
	apiCtx context.Context
	validator Validator
}

// Config represents the configuration of self-update.
type Config struct {
	// APIToken represents GitHub API token. If it's not empty, it will be used for authentication of GitHub API
	APIToken string
	// EnterpriseBaseURL is a base URL of GitHub API. If you want to use this library with GitHub Enterprise,
	// please set "https://{your-organization-address}/api/v3/" to this field.
	EnterpriseBaseURL string
	// EnterpriseUploadURL is a URL to upload stuffs to GitHub Enterprise instance. This is often the same as an API base URL.
	// So if this field is not set and EnterpriseBaseURL is set, EnterpriseBaseURL is also set to this field.
	EnterpriseUploadURL string
	// Validator represents types which enable additional validation of downloaded release.
	Validator Validator
}

func newHTTPClient(ctx context.Context, token string) *http.Client {
	if token == "" {
		return http.DefaultClient
	}
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	return oauth2.NewClient(ctx, src)
}

// NewUpdater creates a new updater instance. It initializes GitHub API client.
// If you set your API token to $GITHUB_TOKEN, the client will use it.
func NewCfgUpdater(config Config, hc *http.Client) (*Updater, error) {
	ctx := context.Background()
	return newUpdater(config, hc,ctx)
}

// NewUpdater creates a new updater instance. It initializes GitHub API client.
// If you set your API token to $GITHUB_TOKEN, the client will use it.
func NewUpdater(config Config) (*Updater, error) {
	token := config.APIToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		token, _ = gitconfig.GithubToken()
	}
	ctx := context.TODO()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	sslclient:= &http.Client{Transport: tr}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, sslclient)
	hc := newHTTPClient(ctx, token)


	return newUpdater(config, hc, ctx)
}

func newUpdater(config Config, hc *http.Client, ctx context.Context) (*Updater, error) {

	if config.EnterpriseBaseURL == "" {
		client := github.NewClient(hc)
		return &Updater{client, ctx, config.Validator}, nil
	}

	u := config.EnterpriseUploadURL
	if u == "" {
		u = config.EnterpriseBaseURL
	}
	client, err := github.NewEnterpriseClient(config.EnterpriseBaseURL, u, hc)
	if err != nil {
		return nil, err
	}
	return &Updater{api: client, apiCtx: ctx, validator: config.Validator}, nil
}


// DefaultUpdater creates a new updater instance with default configuration.
// It initializes GitHub API client with default API base URL.
// If you set your API token to $GITHUB_TOKEN, the client will use it.
func DefaultUpdater() *Updater {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token, _ = gitconfig.GithubToken()
	}
	ctx := context.Background()
	client := newHTTPClient(ctx, token)
	return &Updater{api: github.NewClient(client), apiCtx: ctx}
}
