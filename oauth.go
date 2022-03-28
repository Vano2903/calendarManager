package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	oauth "google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
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
	Ctx          context.Context
	Client       *http.Client
	Config       *oauth2.Config
	OauthToken   *oauth2.Token
	OauthService *oauth.Service
	UserInfo     *oauth.Userinfoplus
	OauthCode    string
	ServerToken  *Token
	//db connection
}

/*
<redirect oauth2 uri callback>?
state=state-token
&code=<oauth generated code>
&scope=
	email
	profile
	https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fcalendar
	https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fcalendar.events
	https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.profile
	https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fuserinfo.email
	openid
&authuser=0
&prompt=consent
*/

//generate the oauth url to let the user authenticate the application
func (o Oauther) GenerateOauthUrl(conn *sql.DB) (string, error) {
	//generaete a random string for the state token
	var expriation time.Time
	var state string
	//repeat until the state token is unique
	for {
		//create a 16 byte random string
		b := make([]byte, 8)
		rand.Read(b)
		state = fmt.Sprintf("%x", b)
		//check if it's in the db already
		var res string
		var t time.Time
		conn.QueryRowContext(o.Ctx, "SELECT * FROM states WHERE value = ?", state).Scan(res, t)
		if res == "" {
			//if not, break the loop
			break
		}
	}

	//5 minute long expiration timestamp s
	expriation = time.Now().Add(time.Minute * 5)
	//insert the state and expiration into the db
	_, err := conn.ExecContext(o.Ctx, "INSERT INTO states (value, expiration) VALUES (?, ?)", state, expriation)
	if err != nil {
		return "", err
	}
	//generate the url with the state string
	return o.Config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

//get the oauth code from the url
func (o *Oauther) GetOauthCodeFromUrl(conn *sql.DB, code, state string) error {
	//check if the state is in the database
	var expriation time.Time
	var stateFromDB string
	err := conn.QueryRowContext(o.Ctx, "SELECT * FROM states WHERE value = ?", state).Scan(&stateFromDB, &expriation)
	if err != nil {
		return err //state not found
	}

	//delete the state from the db
	_, err = conn.ExecContext(o.Ctx, "DELETE FROM states WHERE value = ?", state)
	if err != nil {
		return err
	}

	//check if the state is expired
	if time.Now().After(expriation) {
		return fmt.Errorf("the state is expired")
	}
	o.OauthCode = code
	return nil
}

//get the token pair from the oauth code
//and set the http client
//this function must be ran after GetOauthCodeFromUrl
func (o *Oauther) GetOauthTokenPairFromCode(connection *sql.DB) error {
	//get the token pair from the code
	token, err := o.Config.Exchange(o.Ctx, o.OauthCode)
	if err != nil {
		return err
	}

	//set the token pair
	o.OauthToken = token

	//save the token pair to the db
	_, err = connection.Exec("INSERT INTO googleTokens (googleAccessToken, googleExp, googleRefreshToken) VALUES (?, ?, ?)", o.OauthToken.AccessToken, o.OauthToken.Expiry, o.OauthToken.RefreshToken)
	if err != nil {
		return err
	}
	//generate a new http client
	o.Client = o.Config.Client(context.Background(), token)
	return nil
}

//get the user information from the oauth tokens and set the UserInfo field
//the client must be set, if it's not set run:
//1) GetOauthTokenPairFromCode
//2) SetGoogleTokensFromAccessToken
func (o *Oauther) GetUserInformation() error {
	if o.Client == nil {
		return fmt.Errorf("the client is not set")
	}

	//get the oauth service
	var err error
	o.OauthService, err = oauth.NewService(o.Ctx, option.WithHTTPClient(o.Client))
	if err != nil {
		return fmt.Errorf("unable to create OAuth client: %v", err)
	}

	info, err := oauth.NewUserinfoV2MeService(o.OauthService).Get().Do()
	if err != nil {
		return err
	}
	o.UserInfo = info
	return nil
}

//generate a token pair 64 byte long access and refresh token
//the access token last for 15 minute while the refresh token 1 week
//both of the token are stored in the database, to call this function the
//UserInfo field must not be nil, if it is it will automatically try to retrive it
//!COULD BE A BUG, CANT PROCESS IT NOW THOUGH
func (o *Oauther) GenerateTokensFromUserEmail(connection *sql.DB) error {
	//if the userInfo are not set it will automatically try to get them
	if o.UserInfo == nil {
		err := o.GetUserInformation()
		if err != nil {
			return err
		}
	}

	//generate a random 64 character string and check that's not in the db
	//generate the access token and it expires in 15 minutes
	for {
		//create a 64 byte random string
		b := make([]byte, 32)
		rand.Read(b)
		o.ServerToken.AccessToken = fmt.Sprintf("%x", b)
		o.ServerToken.ExpirationAccessToken = time.Now().Add(time.Minute * 15)
		//check if the access token is already in the database
		found, err := tokenExists(o.ServerToken.AccessToken, connection)
		if err != nil {
			return err
		}

		if !found {
			break
		}
	}

	//same thing of the access token but for the refresh token, the expiration time is longer (7 days)
	for {
		//create a 64 byte random string
		b := make([]byte, 32)
		rand.Read(b)
		o.ServerToken.RefreshToken = fmt.Sprintf("%x", b)
		o.ServerToken.ExpirationRefreshToken = time.Now().Add(time.Hour * 24 * 7)
		//check if the access token is already in the database
		found, err := tokenExists(o.ServerToken.AccessToken, connection)
		if err != nil {
			return err
		}

		if !found {
			break
		}
	}

	//insert the token pair into the database
	accessTokenQuery := "INSERT INTO accessTokens (accToken, accExp) VALUES (?, ?)"
	refreshTokenQuery := "INSERT INTO refreshTokens (refreshToken, refreshExp) VALUES (?, ?)"

	accessResult, err := connection.Exec(accessTokenQuery, o.ServerToken.AccessToken, o.ServerToken.ExpirationAccessToken)
	if err != nil {
		return err
	}

	refreshResult, err := connection.Exec(refreshTokenQuery, o.ServerToken.RefreshToken, o.ServerToken.ExpirationRefreshToken)
	if err != nil {
		return err
	}

	//get the ids of the tokens just inserted
	accessID, _ := accessResult.LastInsertId()
	refreshID, _ := refreshResult.LastInsertId()
	var googleID int
	if err := connection.QueryRow("SELECT ID FROM googleTokens WHERE googleAccessToken = ?", o.OauthToken.AccessToken).Scan(&googleID); err != nil {
		return err
	}

	//insert into tokens on duplicate key update accID = ?, refID = ?, googleID = ?
	relationShipQuery := "INSERT INTO tokens (email, accID, refID, googleID) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE accID = ?, refID = ?, googleID = ?"
	_, err = connection.Exec(relationShipQuery, o.UserInfo.Email, int(accessID), int(refreshID), googleID, int(accessID), int(refreshID), googleID)
	return err
}

//this func will retrive the google tokens from the database given the accessToken
//if the accessToken is not set or expired it will generate a new tokenPair and update the database
func (o *Oauther) SetGoogleTokensFromAccessToken(connection *sql.DB) error {
	//check if the access token is set in the struct
	if o.ServerToken.AccessToken == "" && IsTokenExpired(true, o.ServerToken.AccessToken, connection) {
		if err := o.GenerateNewTokenPairFromRefreshToken(connection); err != nil {
			return err
		}
	}

	//get the google tokens from the access token
	getGoogleTokensFromAccessTokenQuery := `
	SELECT g.ID, g.googleAccessToken, g.googleExp, g.googleRefreshToken, a.accExp 
	FROM googleTokens g 
	JOIN tokens t ON t.googleID = g.ID 
	JOIN accessTokens a ON a.ID = t.accID
	WHERE a.accToken = ?`

	var ID int
	var googleAccessToken, googleRefreshToken string
	var googleExp time.Time

	if err := connection.QueryRowContext(o.Ctx, getGoogleTokensFromAccessTokenQuery, o.ServerToken.AccessToken).Scan(&ID, &googleAccessToken, &googleExp, &googleRefreshToken, &o.ServerToken.ExpirationAccessToken); err != nil {
		return err
	}

	//generate a new oauth2 token pair
	tok := new(oauth2.Token)
	tok.AccessToken = googleAccessToken
	tok.Expiry = googleExp
	tok.RefreshToken = googleRefreshToken
	tok.TokenType = "Bearer"

	//generate a new client, this will automatically autorefresh the token
	//documentation: https://godoc.org/golang.org/x/oauth2#Config.Client
	if o.Client == nil {
		o.Client = o.Config.Client(o.Ctx, tok)
	}

	var err error
	//check if the token is expired
	if !tok.Valid() {
		//get the tokens
		o.OauthToken, err = o.Config.TokenSource(o.Ctx, tok).Token()
		if err != nil {
			return err
		}

		//save the tokens in the database
		updateGoogleTokensQuery := `
		UPDATE googleTokens
		SET googleAccessToken = ?, googleExp = ?, googleRefreshToken = ?
		WHERE ID = ?`
		//update the google tokens
		_, err = connection.Exec(updateGoogleTokensQuery, o.OauthToken.AccessToken, o.OauthToken.Expiry, o.OauthToken.RefreshToken, ID)
	} else {
		o.OauthToken = tok
	}

	return err
}

//if the refreshToken is set this function will generate a new server token pair and automatically
//check if the google tokens are set, if not it will retrive them from the database using the old refreshToken
func (o *Oauther) GenerateNewTokenPairFromRefreshToken(connection *sql.DB) error {
	//check if the refresh token is set in the struct
	if o.ServerToken.RefreshToken == "" {
		return fmt.Errorf("no refresh token")
	}

	//check if the refresh token expiry is set
	if o.ServerToken.ExpirationRefreshToken.IsZero() {
		//if it's not set check the validity of the refresh token from the database
		isExpired := IsTokenExpired(false, o.OauthToken.RefreshToken, connection)
		if isExpired {
			return fmt.Errorf("refresh token is expired, should do the oauth again")
		}
	} else {
		//if it's set just use this one
		//*could have been using just isTokenExpired but it's a database call less
		if time.Now().After(o.OauthToken.Expiry) {
			return fmt.Errorf("refresh token is expired, should do the oauth again")
		}
	}

	//check if the OauthToken field is set
	if o.OauthToken == nil {
		//get the google tokens from the refresh token
		getGoogleTokensFromRefreshTokenQuery := `
		SELECT g.ID, g.googleAccessToken, g.googleExp, g.googleRefreshToken, r.refreshExp
		FROM googleTokens g
		JOIN tokens t ON t.googleID = g.ID
		JOIN refreshTokens r ON r.ID = t.refID
		WHERE r.refreshToken = ?`

		var ID int
		var googleAccessToken, googleRefreshToken string
		var googleExp time.Time
		err := connection.QueryRowContext(o.Ctx, getGoogleTokensFromRefreshTokenQuery, o.OauthToken.RefreshToken).Scan(&ID, &googleAccessToken, &googleExp, &googleRefreshToken, &o.ServerToken.ExpirationRefreshToken)
		if err != nil {
			return err
		}
		//set the google tokens in the struct
		tok := new(oauth2.Token)
		tok.AccessToken = googleAccessToken
		tok.Expiry = googleExp
		tok.RefreshToken = googleRefreshToken
		tok.TokenType = "Bearer"
		o.OauthToken = tok
	}

	//generate new tokens
	return o.GenerateTokensFromUserEmail(connection)
}

//generate a new oauther struct
//the config file is the google api config file (client id, client secret, redirect url, check out the clientgoogle.json.example file)
func NewOauther(credentialFilePath string) (*Oauther, error) {
	//get the config file
	b, err := ioutil.ReadFile(credentialFilePath)
	if err != nil {
		return nil, fmt.Errorf("the file %s coultn't be read", credentialFilePath)
	}

	//create a config pointer from the config file and some google's scope:
	//calendar rw access
	//calendar events rw access
	//user informations and email r-only access
	//the calendar scopes are defined here: https://developers.google.com/calendar/api/guides/auth#OAuth2Scope
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope, calendar.CalendarEventsScope, oauth.UserinfoProfileScope, oauth.UserinfoEmailScope, oauth.PlusMeScope, sheets.DriveFileScope)
	if err != nil {
		return nil, err
	}

	//set the context and the config
	return &Oauther{
		Ctx:    context.Background(),
		Config: config,
		ServerToken: &Token{
			ExpirationAccessToken:  time.Time{},
			AccessToken:            "",
			ExpirationRefreshToken: time.Time{},
			RefreshToken:           "",
		},
	}, nil
}
