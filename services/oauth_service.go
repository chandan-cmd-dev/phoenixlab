package services

import (
	"PhoenixLab/models"
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/beego/beego/v2/client/orm"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// OAuthService manages the single app-level Google OAuth token.
type OAuthService struct{}

func googleRedirectURL() string {
	if v := os.Getenv("GOOGLE_OAUTH_REDIRECT_URL"); v != "" {
		return v
	}
	return "http://localhost:8080/oauth/google/callback"
}

func oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		RedirectURL:  googleRedirectURL(),
		Scopes:       []string{sheets.SpreadsheetsScope},
		Endpoint:     google.Endpoint,
	}
}

// Configured reports whether the app has OAuth client credentials set.
func (s *OAuthService) Configured() bool {
	return os.Getenv("GOOGLE_OAUTH_CLIENT_ID") != "" && os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET") != ""
}

// AuthURL builds the Google consent-screen URL. AccessTypeOffline + prompt=consent
// guarantee a refresh token is issued.
func (s *OAuthService) AuthURL(state string) string {
	return oauthConfig().AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("prompt", "consent"))
}

// Exchange swaps an authorization code for a token.
func (s *OAuthService) Exchange(code string) (*oauth2.Token, error) {
	return oauthConfig().Exchange(context.Background(), code)
}

// SaveToken persists a token as the single connected account. A previous
// refresh token / account email are preserved when the new token omits them.
func (s *OAuthService) SaveToken(tok *oauth2.Token, email string) error {
	o := orm.NewOrm()

	refresh := tok.RefreshToken
	if existing, err := s.LoadToken(); err == nil {
		if refresh == "" {
			refresh = existing.RefreshToken
		}
		if email == "" {
			email = existing.AccountEmail
		}
	}

	if _, err := o.Raw("DELETE FROM google_tokens").Exec(); err != nil {
		return err
	}

	row := &models.GoogleToken{
		AccessToken:  tok.AccessToken,
		RefreshToken: refresh,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
		Scope:        sheets.SpreadsheetsScope,
		AccountEmail: email,
	}
	_, err := o.Insert(row)
	return err
}

func (s *OAuthService) LoadToken() (*models.GoogleToken, error) {
	o := orm.NewOrm()
	var t models.GoogleToken
	if err := o.QueryTable("google_tokens").OrderBy("-Id").Limit(1).One(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *OAuthService) HasToken() bool {
	o := orm.NewOrm()
	c, _ := o.QueryTable("google_tokens").Count()
	return c > 0
}

func (s *OAuthService) Disconnect() error {
	o := orm.NewOrm()
	_, err := o.Raw("DELETE FROM google_tokens").Exec()
	return err
}

func toOAuthToken(t *models.GoogleToken) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		Expiry:       t.Expiry,
	}
}

// TokenSource returns an auto-refreshing token source. Whenever the underlying
// access token is refreshed, the new token is written back to the DB.
func (s *OAuthService) TokenSource() (oauth2.TokenSource, error) {
	gt, err := s.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("no Google account connected")
	}
	base := oauthConfig().TokenSource(context.Background(), toOAuthToken(gt))
	return &persistingTokenSource{base: base, svc: s, last: toOAuthToken(gt)}, nil
}

type persistingTokenSource struct {
	mu   sync.Mutex
	base oauth2.TokenSource
	last *oauth2.Token
	svc  *OAuthService
}

func (p *persistingTokenSource) Token() (*oauth2.Token, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	tok, err := p.base.Token()
	if err != nil {
		return nil, err
	}
	if p.last == nil || tok.AccessToken != p.last.AccessToken {
		_ = p.svc.SaveToken(tok, "")
		p.last = tok
	}
	return tok, nil
}
