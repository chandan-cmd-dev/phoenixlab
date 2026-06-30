package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SheetsClient struct {
	svc *sheets.Service
}

type SheetMeta struct {
	Title string
	Tabs  []string
}

type CellUpdate struct {
	Row   int
	Col   int
	Value string
}

func NewSheetsClient() (*SheetsClient, error) {
	oauthSvc := &OAuthService{}
	ts, err := oauthSvc.TokenSource()
	if err != nil {
		return nil, err
	}
	svc, err := sheets.NewService(context.Background(), option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return &SheetsClient{svc: svc}, nil
}

var spreadsheetIDRegex = regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9-_]+)`)

func ExtractSpreadsheetID(input string) string {
	input = strings.TrimSpace(input)
	if m := spreadsheetIDRegex.FindStringSubmatch(input); len(m) == 2 {
		return m[1]
	}
	return input
}

func (c *SheetsClient) GetMeta(spreadsheetID string) (*SheetMeta, error) {
	ss, err := c.svc.Spreadsheets.Get(spreadsheetID).
		Fields("properties.title", "sheets.properties.title").Do()
	if err != nil {
		return nil, err
	}
	meta := &SheetMeta{Title: ss.Properties.Title}
	for _, sh := range ss.Sheets {
		meta.Tabs = append(meta.Tabs, sh.Properties.Title)
	}
	return meta, nil
}

func (c *SheetsClient) GetRows(spreadsheetID, tabName string) ([][]string, error) {
	resp, err := c.svc.Spreadsheets.Values.Get(spreadsheetID, escapeTab(tabName)).
		ValueRenderOption("FORMATTED_VALUE").Do()
	if err != nil {
		return nil, err
	}
	rows := make([][]string, 0, len(resp.Values))
	for _, r := range resp.Values {
		row := make([]string, 0, len(r))
		for _, cell := range r {
			row = append(row, fmt.Sprintf("%v", cell))
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (c *SheetsClient) WriteCells(spreadsheetID, tabName string, updates []CellUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	data := make([]*sheets.ValueRange, 0, len(updates))
	for _, u := range updates {
		data = append(data, &sheets.ValueRange{
			Range:  a1Cell(tabName, u.Row, u.Col),
			Values: [][]interface{}{{u.Value}},
		})
	}
	_, err := c.svc.Spreadsheets.Values.BatchUpdate(spreadsheetID, &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "RAW",
		Data:             data,
	}).Do()
	return err
}

func colToLetters(col int) string {
	col++
	s := ""
	for col > 0 {
		col--
		s = string(rune('A'+col%26)) + s
		col /= 26
	}
	return s
}

func escapeTab(tab string) string {
	return "'" + strings.ReplaceAll(tab, "'", "''") + "'"
}

func a1Cell(tab string, row, col int) string {
	return fmt.Sprintf("%s!%s%d", escapeTab(tab), colToLetters(col), row+1)
}
