package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	oauth "google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}

	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("premi l'url e, dopo aver verificato le varie cose copia il valore del parametro 'code': \n\n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {
	ctx := context.Background()
	b, err := ioutil.ReadFile("clientgoogle.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope, calendar.CalendarEventsScope, oauth.UserinfoProfileScope, oauth.UserinfoEmailScope, oauth.PlusMeScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	s, err := oauth.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to create OAuth client: %v", err)
	}

	info, err := oauth.NewUserinfoV2MeService(s).Get().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve user info: %v", err)
	}
	fmt.Println("======INFO GENERALI======\n")

	fmt.Println("======UTENTE INFO 1======")
	fmt.Printf("email %q\n", info.Email)
	fmt.Printf("family name %q\n", info.FamilyName)
	fmt.Printf("gender %q\n", info.Gender)
	fmt.Printf("given name %q\n", info.GivenName)
	fmt.Printf("hosted domain %q\n", info.Hd)
	fmt.Printf("ID %q\n", info.Id)
	fmt.Printf("link %q\n", info.Link)
	fmt.Printf("locale %q\n", info.Locale)
	fmt.Printf("name %q\n", info.Name)
	fmt.Printf("picure %q\n", info.Picture)
	fmt.Printf("verified %v\n", *info.VerifiedEmail)

	info2, err := s.Tokeninfo().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve user info: %v", err)
	}
	fmt.Println("======UTENTE INFO 2 (inutile)======")
	fmt.Printf("access_type %q\n", info2.AccessType)
	fmt.Printf("audience %q\n", info2.Audience)
	fmt.Printf("email %q\n", info2.Email)
	fmt.Printf("verified_email %v\n", info2.VerifiedEmail)
	fmt.Printf("email_verified %v\n", info2.EmailVerified)
	fmt.Printf("expires_in %q\n", info2.ExpiresIn)
	fmt.Printf("issued_at %q\n", info2.IssuedAt)
	fmt.Printf("issued_to %q\n", info2.IssuedTo)
	fmt.Printf("issuer %q\n", info2.Issuer)
	fmt.Printf("nonce %q\n", info2.Nonce)
	fmt.Printf("scope %q\n", info2.Scope)
	fmt.Printf("user_id %q\n", info2.UserId)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}
	srv.UserAgent = "calendar-manager-v1"
	fmt.Println("\n\n======CALENDARS JSON======\n")
	// fmt.Println("======EVENTS======")
	// t := time.Now().Format(time.RFC3339)
	// events, err := srv.Events.List("primary").ShowDeleted(true).SingleEvents(true).TimeMin(t).OrderBy("startTime").Do() //SingleEvents(true).TimeMin(t).MaxResults(10).
	// if err != nil {
	// 	log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	// }

	// fmt.Println("Upcoming events:")
	// if len(events.Items) == 0 {
	// 	fmt.Println("No upcoming events found.")
	// 	return
	// }

	// for _, item := range events.Items {
	// 	date := item.Start.DateTime
	// 	if date == "" {
	// 		date = item.Start.Date
	// 	}
	// 	fmt.Printf("%v (%v)\n", item.Summary, date)
	// }

	//!googleapi.IsNotModified(err)
	idk, err := srv.CalendarList.List().Do()
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar List: %v", err)
	}

	fmt.Println("======CALENDAR LIST FIELDS======")

	fmt.Println("calendar list etag", idk.Etag)
	fmt.Println("calendar list next page token", idk.NextPageToken)
	fmt.Println("calendar list next sync token", idk.NextSyncToken)

	fmt.Println("======REMINDERS======")

	for i, item := range idk.Items {
		// fmt.Printf("%v (%v)\n", item.Summary, item.Id)
		fmt.Printf("======ELEMENT %d======\n", i)

		fmt.Println("======CONFERENCE======")
		fmt.Println("Conference proprieties solution:", item.ConferenceProperties.AllowedConferenceSolutionTypes)

		fmt.Println("\n======REMINDERS======")
		if item.DefaultReminders != nil {
			for _, reminder := range item.DefaultReminders {
				fmt.Println("event reminder method:", reminder.Method)
				fmt.Println("event reminder minutes:", reminder.Minutes)
			}
		} else {
			fmt.Println("event reminder:", "NONE")
		}

		fmt.Println("\n======NOTIFICATION======")
		if item.NotificationSettings != nil {
			for _, notification := range item.NotificationSettings.Notifications {
				fmt.Println("notification method:", notification.Method)
				fmt.Println("notification Type:", notification.Type)
			}
		} else {
			fmt.Println("event notification:", "NONE")
		}

		fmt.Println("\n======ITEM FIELD======")
		fmt.Println("accessRole:", item.AccessRole)
		fmt.Println("background color:", item.BackgroundColor)
		fmt.Println("colorId:", item.ColorId)
		fmt.Println("deleted:", item.Deleted)
		fmt.Println("description:", item.Description)
		fmt.Println("etag:", item.Etag)
		fmt.Println("foreground color:", item.ForegroundColor)
		fmt.Println("hidden:", item.Hidden)
		fmt.Println("king:", item.Kind)
		fmt.Println("location:", item.Location)
		fmt.Println("primary:", item.Primary)
		fmt.Println("selected:", item.Selected)
		fmt.Println("summary:", item.Summary)
		fmt.Println("summaryOverride:", item.SummaryOverride)
		fmt.Println("timeZone:", item.TimeZone)
		fmt.Println()
	}
}
