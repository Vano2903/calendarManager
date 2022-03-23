package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	oauth "google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

type Token struct {
	AccessToken            string
	ExpirationAccessToken  time.Time
	RefreshToken           string
	ExpirationRefreshToken time.Time
}

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
func (o *Oauther) GenerateOauthUrl(conn *sql.Conn) (string, error) {
	//generaete a random string for the state token
	var expriation time.Time
	var state string
	//repeat until the state token is unique
	for {
		//create a 16 byte random string
		b := make([]byte, 8)
		rand.Read(b)
		state = fmt.Sprintf("%x", b)
		//1 minute long expiration timestamp
		expriation = time.Now().Add(time.Minute * 5)
		var res string
		var t time.Time
		//check if it's in the db already
		conn.QueryRowContext(o.Ctx, "SELECT * FROM oauth WHERE state = ?", state).Scan(res, t)
		if res == "" {
			//if not, break the loop
			break
		}
	}

	//insert the state and expiration into the db
	_, err := conn.ExecContext(o.Ctx, "INSERT INTO states (value, expiration) VALUES (?, ?)", state, expriation)
	if err != nil {
		return "", err
	}
	//generate the url with the state string
	return o.Config.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

//get the oauth code from the url
func (o *Oauther) GetOauthCodeFromUrl(conn *sql.Conn, oauthUrl string) error {
	//decode the url encoded url string
	//%2F => /
	oauthUrlDecoded, err := url.QueryUnescape(oauthUrl)
	if err != nil {
		return err
	}

	//convert the url parameters to a map
	vals, err := url.ParseQuery(oauthUrlDecoded)
	if err != nil {
		return err
	}

	state := vals.Get("state")

	//check if the state is in the database
	var expriation time.Time
	var stateFromDB string
	err = conn.QueryRowContext(o.Ctx, "SELECT * FROM states WHERE value = ?", state).Scan(&stateFromDB, &expriation)
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
	o.OauthCode = vals.Get("code")
	return nil
}

//get the token pair from the oauth code
//and set the http client
func (o *Oauther) GetOauthTokenPairFromCode() error {
	//get the token pair from the code
	token, err := o.Config.Exchange(o.Ctx, o.OauthCode)
	if err != nil {
		return err
	}

	//set the token pair
	o.OauthToken = token
	o.Client = o.Config.Client(context.Background(), token)
	return nil
}

//get the user information from the oauth tokens and set the UserInfo field
func (o *Oauther) GetUserInformation() error {
	//get the oauth service
	var err error
	o.OauthService, err = oauth.NewService(o.Ctx, option.WithHTTPClient(o.Client))
	if err != nil {
		log.Fatalf("Unable to create OAuth client: %v", err)
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
//UserInfo field must not be nil
func (o *Oauther) GenerateTokensFromUserEmail(connection *sql.DB) error {
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

	//relationship between the tokens and the email
	relationShipQuery := "INSERT INTO tokens (email, accID, refID) VALUES (?, ?, ?)"
	_, err = connection.Exec(relationShipQuery, o.UserInfo.Email, accessID, refreshID)

	return err
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
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope, calendar.CalendarEventsScope, oauth.UserinfoProfileScope, oauth.UserinfoEmailScope, oauth.PlusMeScope)
	if err != nil {
		return nil, err
	}

	//set the context and the config
	return &Oauther{
		Ctx:    context.Background(),
		Config: config,
	}, nil
}

//check if a token is already in the database (either access or refresh token)
func tokenExists(token string, connection *sql.DB) (bool, error) {
	query := "SELECT * FROM tokens t JOIN accessTokens a ON a.ID = t.accID JOIN refreshTokens r ON r.ID = t.refID WHERE a.accToken = ? OR r.refreshToken = ?"

	var found bool
	err := connection.QueryRow(query, token, token).Scan(&found)
	if err != nil {
		//if the error is ErrNoRows then the token is not in the database
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}
