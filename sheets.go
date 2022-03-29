package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheeter struct {
	Ctx           context.Context
	Client        *http.Client
	SheetsService *sheets.Service
	SheetUri      string
	SheetID       string
	Events        []Event
}

type Event struct {
	Title       string        `json:"title"`
	Description string        `json:"description"`
	StartTime   time.Time     `json:"startTime"`
	Duration    time.Duration `json:"duration"`
	EndTime     time.Time     `json:"endTime"`
	Color       string        `json:"color"`
}

//create a new sheets and return the sheet id and the url
//it also saves the sheet id with the email of the owner in the database
func (s *Sheeter) CreateNewSheet(connection *sql.DB, email, sheetName string) error {
	sheet := &sheets.Spreadsheet{
		Properties: &sheets.SpreadsheetProperties{
			Title: sheetName,
		},
	}
	created, err := s.SheetsService.Spreadsheets.Create(sheet).Do()
	if err != nil {
		return err
	}
	//set the headers of the sheet
	cellRange := "A1:E1"
	valueRange := &sheets.ValueRange{
		Range: cellRange,
		Values: [][]interface{}{
			{"Titolo", "Descrizione", "Data e ora d'inizio", "Durata", "Colore"},
		},
	}

	call := s.SheetsService.Spreadsheets.Values.Update(created.SpreadsheetId, cellRange, valueRange)
	call.ValueInputOption("RAW")
	if _, err := call.Do(); err != nil {
		return err
	}
	//save the sheet id with the email of the owner in the database
	//on duplicate key update the sheet id
	_, err = connection.Exec("INSERT INTO sheets (emailOwner, sheetID) VALUES (?, ?) ON DUPLICATE KEY UPDATE sheetID = ?", email, created.SpreadsheetId, created.SpreadsheetId)
	if err != nil {
		return err
	}
	s.SheetID, s.SheetUri = created.SpreadsheetId, created.SpreadsheetUrl
	return nil
}

//get the sheet url from the owner email
func (s *Sheeter) GetSheetUrlFromEmail(connection *sql.DB, email string) error {
	var sheetID string
	err := connection.QueryRow("SELECT sheetID FROM sheets WHERE emailOwner = ?", email).Scan(&sheetID)
	if err != nil {
		return err
	}
	//get the url from the sheet id
	sheet, err := s.SheetsService.Spreadsheets.Get(sheetID).Do()
	if err != nil {
		return err
	}
	s.SheetID, s.SheetUri = sheetID, sheet.SpreadsheetUrl
	return nil
}

//check if the user has a google sheets linked to his email
func (s Sheeter) DoesUserOwnSheet(connection *sql.DB, email string) (bool, error) {
	var count int
	err := connection.QueryRow("SELECT COUNT(*) FROM sheets WHERE emailOwner = ?", email).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (s *Sheeter) GetEventsFromSheet(connection *sql.DB, email string) error {
	if s.SheetID == "" {
		//get the sheet id
		err := connection.QueryRow("SELECT sheetID FROM sheets WHERE emailOwner = ?", email).Scan(&s.SheetID)
		if err != nil {
			return err
		}
	}

	//get the sheet data
	resp, err := s.SheetsService.Spreadsheets.Values.Get(s.SheetID, "A2:E").Do()
	if err != nil {
		return err
	}
	//check if there are events in the sheet
	if len(resp.Values) == 0 {
		return nil
	}
	//get the events
	s.Events = []Event{}
	for i, row := range resp.Values {
		// start := strings.Split(row[2].(string), " ")
		start, err := time.Parse("2006-01-02 15:04", row[2].(string))
		if err != nil {
			return fmt.Errorf("was unable to parse the datetime at row C%d: %v", i+2, err)
		}

		duration := strings.Split(row[3].(string), ":")
		d, err := time.ParseDuration(fmt.Sprintf("%sh%sm", duration[0], duration[1]))
		if err != nil {
			return fmt.Errorf("was unable to parse the duration at row D%d: %v", i+2, err)
		}
		event := Event{
			Title:       row[0].(string),
			Description: row[1].(string),
			StartTime:   start,
			Duration:    d,
			EndTime:     start.Add(d),
			Color:       row[4].(string),
		}
		s.Events = append(s.Events, event)
	}
	return nil
}

//given the oauther client it will set the sheets service
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
