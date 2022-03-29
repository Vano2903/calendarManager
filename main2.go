package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

var (
	handler *HandlerHttp
)

func init() {
	var err error
	if handler, err = NewHandlerHttp(googleCredentialsFile); err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := mux.NewRouter()
	r.Use(handler.loggerMiddleware)

	//serve static files
	r.PathPrefix("/static").Handler(http.FileServer(http.Dir("./static/")))
	//!need to know the pages first
	// r.HandleFunc("/oauth", oauthGooglePageHandler).Methods("GET")

	sheets := r.PathPrefix("/sheets").Subrouter()
	sheets.Use(handler.AccessTokenMiddleware)
	sheets.HandleFunc("/new", handler.CreateSheetHandler).Methods("GET")
	sheets.HandleFunc("/events", handler.GetEventsHandler).Methods("GET")
	
	oauth := r.PathPrefix("/auth").Subrouter()
	oauth.HandleFunc("/google/login", handler.OauthGoogleHandler).Methods("GET")
	oauth.HandleFunc("/google/callback", handler.OauthGoogleCallbackHandler).Methods("GET")

	fmt.Println("listening on port 8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal(err)
	}
}

// func getUserInfoHandler(w http.ResponseWriter, r *http.Request) {
// 	conn, err := connectToDB()
// 	if err != nil {
// 		fmt.Printf("Error connecting to DB: %v", err)
// 		resp.Error(w, http.StatusInternalServerError, "Error connecting to DB")
// 		return
// 	}
// 	defer conn.Close()

// 	//get the token pair from cookies
// 	//if it's not set, it should refresh the request because this function pass
// 	for _, cookie := range r.Cookies() {
// 		switch cookie.Name {
// 		case "accessToken":
// 			oauther.ServerToken.AccessToken = cookie.Value
// 		case "refreshToken":
// 			oauther.ServerToken.RefreshToken = cookie.Value
// 		}
// 	}

// 	err = oauther.SetGoogleTokensFromAccessToken(conn)
// 	if err != nil {
// 		fmt.Printf("error in oauther function, setting google tokens from access token: %v", err)
// 		resp.Errorf(w, http.StatusInternalServerError, "Error setting google tokens from access token: %v", err)
// 		return
// 	}

// 	//get the user info
// 	err = oauther.GetUserInformation()
// 	if err != nil {
// 		fmt.Printf("error in oauther function, getting user info: %v", err)
// 		resp.Errorf(w, http.StatusInternalServerError, "Error getting user info: %v", err)
// 		return
// 	}

// 	//marshal the user info
// 	json, err := json.Marshal(oauther.UserInfo)
// 	if err != nil {
// 		fmt.Printf("error in oauther function, marshaling user info: %v", err)
// 		resp.Errorf(w, http.StatusInternalServerError, "Error marshaling user info: %v", err)
// 		return
// 	}
// 	resp.SuccessJson(w, http.StatusOK, "Successfully got user info", json)
// }

// func main() {
// 	ctx := context.Background()
// 	b, err := ioutil.ReadFile("clientgoogle.json")
// 	if err != nil {
// 		log.Fatalf("Unable to read client secret file: %v", err)
// 	}

// 	// If modifying these scopes, delete your previously saved token.json.
// 	config, err := google.ConfigFromJSON(b, calendar.CalendarScope, calendar.CalendarEventsScope, oauth.UserinfoProfileScope, oauth.UserinfoEmailScope, oauth.PlusMeScope)
// 	if err != nil {
// 		log.Fatalf("Unable to parse client secret file to config: %v", err)
// 	}
// 	client := getClient(config)

// 	s, err := oauth.NewService(ctx, option.WithHTTPClient(client))
// 	if err != nil {
// 		log.Fatalf("Unable to create OAuth client: %v", err)
// 	}

// 	info, err := oauth.NewUserinfoV2MeService(s).Get().Do()
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve user info: %v", err)
// 	}
// 	fmt.Println("======INFO GENERALI======\n")

// 	fmt.Println("======UTENTE INFO 1======")
// 	fmt.Printf("email %q\n", info.Email)
// 	fmt.Printf("family name %q\n", info.FamilyName)
// 	fmt.Printf("gender %q\n", info.Gender)
// 	fmt.Printf("given name %q\n", info.GivenName)
// 	fmt.Printf("hosted domain %q\n", info.Hd)
// 	fmt.Printf("ID %q\n", info.Id)
// 	fmt.Printf("link %q\n", info.Link)
// 	fmt.Printf("locale %q\n", info.Locale)
// 	fmt.Printf("name %q\n", info.Name)
// 	fmt.Printf("picure %q\n", info.Picture)
// 	fmt.Printf("verified %v\n", *info.VerifiedEmail)

// 	info2, err := s.Tokeninfo().Do()
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve user info: %v", err)
// 	}
// 	fmt.Println("======UTENTE INFO 2 (inutile)======")
// 	fmt.Printf("access_type %q\n", info2.AccessType)
// 	fmt.Printf("audience %q\n", info2.Audience)
// 	fmt.Printf("email %q\n", info2.Email)
// 	fmt.Printf("verified_email %v\n", info2.VerifiedEmail)
// 	fmt.Printf("email_verified %v\n", info2.EmailVerified)
// 	fmt.Printf("expires_in %q\n", info2.ExpiresIn)
// 	fmt.Printf("issued_at %q\n", info2.IssuedAt)
// 	fmt.Printf("issued_to %q\n", info2.IssuedTo)
// 	fmt.Printf("issuer %q\n", info2.Issuer)
// 	fmt.Printf("nonce %q\n", info2.Nonce)
// 	fmt.Printf("scope %q\n", info2.Scope)
// 	fmt.Printf("user_id %q\n", info2.UserId)

// 	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve Calendar client: %v", err)
// 	}
// 	srv.UserAgent = "calendar-manager-v1"
// 	fmt.Println("\n\n======CALENDARS JSON======\n")
// 	// fmt.Println("======EVENTS======")
// 	// t := time.Now().Format(time.RFC3339)
// 	// events, err := srv.Events.List("primary").ShowDeleted(true).SingleEvents(true).TimeMin(t).OrderBy("startTime").Do() //SingleEvents(true).TimeMin(t).MaxResults(10).
// 	// if err != nil {
// 	// 	log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
// 	// }

// 	// fmt.Println("Upcoming events:")
// 	// if len(events.Items) == 0 {
// 	// 	fmt.Println("No upcoming events found.")
// 	// 	return
// 	// }

// 	// for _, item := range events.Items {
// 	// 	date := item.Start.DateTime
// 	// 	if date == "" {
// 	// 		date = item.Start.Date
// 	// 	}
// 	// 	fmt.Printf("%v (%v)\n", item.Summary, date)
// 	// }

// 	//!googleapi.IsNotModified(err)
// 	idk, err := srv.CalendarList.List().Do()
// 	if err != nil {
// 		log.Fatalf("Unable to retrieve Calendar List: %v", err)
// 	}

// 	fmt.Println("======CALENDAR LIST FIELDS======")

// 	fmt.Println("calendar list etag", idk.Etag)
// 	fmt.Println("calendar list next page token", idk.NextPageToken)
// 	fmt.Println("calendar list next sync token", idk.NextSyncToken)

// 	fmt.Println("======REMINDERS======")

// 	for i, item := range idk.Items {
// 		// fmt.Printf("%v (%v)\n", item.Summary, item.Id)
// 		fmt.Printf("======ELEMENT %d======\n", i)

// 		fmt.Println("======CONFERENCE======")
// 		fmt.Println("Conference proprieties solution:", item.ConferenceProperties.AllowedConferenceSolutionTypes)

// 		fmt.Println("\n======REMINDERS======")
// 		if item.DefaultReminders != nil {
// 			for _, reminder := range item.DefaultReminders {
// 				fmt.Println("event reminder method:", reminder.Method)
// 				fmt.Println("event reminder minutes:", reminder.Minutes)
// 			}
// 		} else {
// 			fmt.Println("event reminder:", "NONE")
// 		}

// 		fmt.Println("\n======NOTIFICATION======")
// 		if item.NotificationSettings != nil {
// 			for _, notification := range item.NotificationSettings.Notifications {
// 				fmt.Println("notification method:", notification.Method)
// 				fmt.Println("notification Type:", notification.Type)
// 			}
// 		} else {
// 			fmt.Println("event notification:", "NONE")
// 		}

// 		fmt.Println("\n======ITEM FIELD======")
// 		fmt.Println("accessRole:", item.AccessRole)
// 		fmt.Println("background color:", item.BackgroundColor)
// 		fmt.Println("colorId:", item.ColorId)
// 		fmt.Println("deleted:", item.Deleted)
// 		fmt.Println("description:", item.Description)
// 		fmt.Println("etag:", item.Etag)
// 		fmt.Println("foreground color:", item.ForegroundColor)
// 		fmt.Println("hidden:", item.Hidden)
// 		fmt.Println("king:", item.Kind)
// 		fmt.Println("location:", item.Location)
// 		fmt.Println("primary:", item.Primary)
// 		fmt.Println("selected:", item.Selected)
// 		fmt.Println("summary:", item.Summary)
// 		fmt.Println("summaryOverride:", item.SummaryOverride)
// 		fmt.Println("timeZone:", item.TimeZone)
// 		fmt.Println()
// 	}
// }
