package main

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Calender struct {
	Ctx             context.Context
	Client          *http.Client
	CalendarService *calendar.Service
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
