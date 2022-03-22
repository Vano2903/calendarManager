package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	oauth "google.golang.org/api/oauth2/v1"
)

/*
functions:
=====OAUTH
GenerateOauthUrl
GetOauthCodeFromUrl
GetOauthTokenPairFromCode
GetUserInformation
=====SELF TOKENS
GenerateTokensFromUserEmail
GenerateAccessTokenFromRefreshToken
SaveTokensToDb -> email, oauthAccess, oauthRefresh, serverAcces, serverRefresh
*/
type Oauther struct {
	ctx          context.Context
	config       *oauth2.Config
	oauthToken   *oauth2.Token
	oauthService *oauth.Service
	Email        string
	serverToken  *oauth2.Token
	//db connection
}

func NewOauther(credentialFilePath string) (*Oauther, error) {
	b, err := ioutil.ReadFile(credentialFilePath)
	if err != nil {
		return nil, fmt.Errorf("the file %s coultn't be read", credentialFilePath)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarScope, calendar.CalendarEventsScope, oauth.UserinfoProfileScope, oauth.UserinfoEmailScope, oauth.PlusMeScope)
	if err != nil {
		return nil, err
	}

	return &Oauther{
		ctx:    context.Background(),
		config: config,
	}, nil
}
