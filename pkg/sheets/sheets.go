package sheets

import (
	"context"
	"fmt"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
)

type SheetsService struct {
	Service *sheets.Service
	SheetID string
}

func NewSheetsService(credentialsPath string, sheetID string) (*SheetsService, error) {
	ctx := context.Background()
	b, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}

	client := config.Client(ctx)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	return &SheetsService{
		Service: srv,
		SheetID: sheetID,
	}, nil
}

func (s *SheetsService) ReadSheet(sheetName string) ([][]interface{}, error) {
	rangeData := fmt.Sprintf("%s!A1:Z1000", sheetName)

	fmt.Println("Reading sheet:", sheetName)
	fmt.Println("Sheet ID:", s.SheetID)

	resp, err := s.Service.Spreadsheets.Values.Get(s.SheetID, rangeData).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}
	return resp.Values, nil
}

func (s *SheetsService) UpdateSheet(sheetName string, data [][]interface{}) error {
	rangeData := fmt.Sprintf("%s!A1", sheetName)
	valueRange := &sheets.ValueRange{
		Values: data,
	}

	_, err := s.Service.Spreadsheets.Get(s.SheetID).Do()
	if err != nil {
		return fmt.Errorf("unable to retrieve spreadsheet: %v", err)
	}

	// Check that the sheetName exists in the spreadsheet
	_, err = s.Service.Spreadsheets.Values.Get(s.SheetID, rangeData).Do()
	if err != nil {
		// Try to create the sheetName if it doesn't exist
		_, err = s.Service.Spreadsheets.BatchUpdate(s.SheetID, &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{
				&sheets.Request{
					AddSheet: &sheets.AddSheetRequest{
						Properties: &sheets.SheetProperties{
							Title: sheetName,
						},
					},
				},
			},
		}).Do()
		if err != nil {
			return fmt.Errorf("unable to create sheet: %v", err)
		}

		return fmt.Errorf("unable to retrieve data from sheet: %v", err)
	}

	_, err = s.Service.Spreadsheets.Values.Update(s.SheetID, rangeData, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return fmt.Errorf("unable to update data in sheet: %v", err)
	}

	return nil
}
