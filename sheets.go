package main

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheeter struct {
	Ctx           context.Context
	Client        *http.Client
	SheetsService *sheets.Service
}

//create a new sheets and return the sheet id and the url
func (s Sheeter) CreateNewSheet(sheetName string) (string, string, error) {
	sheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: sheetName,
		},
	}
	created, err := s.SheetsService.Spreadsheets.Create(sheet).Do()
	if err != nil {
		return "", "", err
	}
	return created.SpreadsheetId, created.SpreadsheetUrl, nil
}

func NewSheeter(client *http.Client) (*Sheeter, error) {
	if client == nil {
		return nil, fmt.Errorf("client is nil")
	}
	s, err := sheets.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}
	return &Sheeter{
		Ctx:           context.Background(),
		Client:        client,
		SheetsService: s,
	}, nil
}
