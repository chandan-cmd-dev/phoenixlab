package models

import (
	"time"
)

// GoogleToken stores the single app-level OAuth token used to talk to the
// Google Sheets API. The OAuth service keeps exactly one row in this table.
type GoogleToken struct {
	Id           int       `orm:"auto;pk" json:"id"`
	AccessToken  string    `orm:"type(text)" json:"-"`
	RefreshToken string    `orm:"null;type(text)" json:"-"`
	TokenType    string    `orm:"null;size(50)" json:"token_type"`
	Expiry       time.Time `orm:"null;type(timestamptz)" json:"expiry"`
	Scope        string    `orm:"null;type(text)" json:"scope"`
	AccountEmail string    `orm:"null;size(200)" json:"account_email"`
	CreatedAt    time.Time `orm:"auto_now_add;type(timestamptz)" json:"created_at"`
	UpdatedAt    time.Time `orm:"auto_now;type(timestamptz)" json:"updated_at"`
}

func (t *GoogleToken) TableName() string {
	return "google_tokens"
}
