package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calender struct {
	Ctx             context.Context
	Client          *http.Client
	CalendarService *calendar.Service
	CalendarID      string
}

//check if the user has a calendar id set in the database, if so it will also set it
func (c *Calender) DoesUserOwnCalendar(connection *sql.DB, email string) (bool, error) {
	var calendarID string
	err := connection.QueryRow("SELECT calendarID FROM calendars WHERE emailOwner = ?", email).Scan(&calendarID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	c.CalendarID = calendarID
	return true, nil
}

//create a new calendar, this function check if the user already has a calendar,
//if so it will return an error, if he doesnt then it will create it and set it on the database
func (c *Calender) CreateNewCalendar(connection *sql.DB, email, name string) error {
	if itDoes, err := c.DoesUserOwnCalendar(connection, email); err != nil {
		return err
	} else if itDoes {
		return fmt.Errorf("user already has a calendar")
	}
	//create a new calendar
	calendar := &calendar.Calendar{
		Summary:     name,
		Description: "Calendar created by the Calendar Manager Team",
		Location:    "Rome",
	}
	megaSuperCalendar, err := c.CalendarService.Calendars.Insert(calendar).Do()
	if err != nil {
		return err
	}

	c.CalendarID = megaSuperCalendar.Id
	fmt.Println("Calendar ID: ", c.CalendarID)
	//save the calendar id with the email of the owner in the database
	//on duplicate key update the calendar id
	_, err = connection.Exec("INSERT INTO calendars (emailOwner, calendarID) VALUES (?, ?)", email, megaSuperCalendar.Id)
	return err
}

//update the calendar with all the events (it will delete all the events and then add them again)
func (c *Calender) UpdateCalendar(events []Event) error {
	//delete all the events from the calendar
	eventsToDelete, err := c.CalendarService.Events.List(c.CalendarID).ShowDeleted(true).Do()
	if err != nil {
		return err
	}
	for _, event := range eventsToDelete.Items {
		err := c.CalendarService.Events.Delete(c.CalendarID, event.Id).Do()
		if err != nil {
			return err
		}
	}

	// create the events
	for _, event := range events {
		err := c.CreateEvent(event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Calender) CreateEvent(event Event) error {
	eventToCreate := &calendar.Event{
		Summary:     event.Title,
		Description: event.Description,
		// Location:    event.Location,
		Start: &calendar.EventDateTime{
			DateTime: event.StartTime.Format(time.RFC3339),
			TimeZone: "Europe/Rome",
		},
		End: &calendar.EventDateTime{
			DateTime: event.EndTime.Format(time.RFC3339),
			TimeZone: "Europe/Rome",
		},
		// ColorId: event.Color,
	}
	_, err := c.CalendarService.Events.Insert(c.CalendarID, eventToCreate).Do()
	if err != nil {
		return err
	}
	return nil
}

func NewCalender(client *http.Client) (*Calender, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	s, err := calendar.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return &Calender{
		Ctx:             context.Background(),
		Client:          client,
		CalendarService: s,
	}, nil
}
