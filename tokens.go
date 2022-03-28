package main

import (
	"database/sql"
	"fmt"
	"time"
)

type Token struct {
	AccessToken            string    `json:"access_token"`
	ExpirationAccessToken  time.Time `json:"expiration_access_token"`
	RefreshToken           string    `json:"refresh_token"`
	ExpirationRefreshToken time.Time `json:"expiration_refresh_token"`
}

//check if a token is already in the database (either access or refresh token)
func tokenExists(token string, connection *sql.DB) (bool, error) {
	query := "SELECT * FROM tokens t JOIN accessTokens a ON a.ID = t.accID JOIN refreshTokens r ON r.ID = t.refID WHERE a.accToken = ? OR r.refreshToken = ?"

	var found bool
	err := connection.QueryRow(query, token, token).Scan(&found)
	if err != nil {
		//if the error is ErrNoRows then the token is not in the database
		if err == sql.ErrNoRows {
			fmt.Printf("token %s not found", token)
			return false, nil
		}
		return false, err
	}

	return true, nil
}

//check if the token is expired (must define the type of token)
func IsTokenExpired(isAccessToken bool, token string, connection *sql.DB) bool {
	var query string
	if isAccessToken {
		query = "SELECT accExp FROM accessTokens WHERE accToken = ?"
	} else {
		query = "SELECT refreshExp FROM refreshTokens WHERE refreshToken = ?"
	}

	var exp time.Time
	err := connection.QueryRow(query, token).Scan(&exp)
	if err != nil {
		return true
	}

	//check if the expiration time is "before" the current time
	return exp.Before(time.Now())
}
