package handlers_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"net/http"
	"strconv"
	"testing"
)

func createReport(client *http.Client, report *handlers.Report) (*handlers.Report, error) {
	var r handlers.Report
	err := create(client, report, &r, "/report/", "report")
	return &r, err
}

func getReport(client *http.Client, reportid int64) (*handlers.Report, error) {
	var r handlers.Report
	err := read(client, &r, "/report/"+strconv.FormatInt(reportid, 10), "report")
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func getReports(client *http.Client) (*handlers.ReportList, error) {
	var rl handlers.ReportList
	err := read(client, &rl, "/report/", "reports")
	if err != nil {
		return nil, err
	}
	return &rl, nil
}

func updateReport(client *http.Client, report *handlers.Report) (*handlers.Report, error) {
	var r handlers.Report
	err := update(client, report, &r, "/report/"+strconv.FormatInt(report.ReportId, 10), "report")
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func deleteReport(client *http.Client, r *handlers.Report) error {
	err := remove(client, "/report/"+strconv.FormatInt(r.ReportId, 10), "report")
	if err != nil {
		return err
	}
	return nil
}

func TestCreateReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].reports); i++ {
			orig := data[0].reports[i]
			r := d.reports[i]

			if r.ReportId == 0 {
				t.Errorf("Unable to create report: %+v", r)
			}
			if r.Name != orig.Name {
				t.Errorf("Name doesn't match")
			}
			if r.Lua != orig.Lua {
				t.Errorf("Lua doesn't match")
			}
		}
	})
}

func TestGetReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].reports); i++ {
			orig := data[0].reports[i]
			curr := d.reports[i]

			r, err := getReport(d.clients[orig.UserId], curr.ReportId)
			if err != nil {
				t.Fatalf("Error fetching reports: %s\n", err)
			}
			if r.Name != orig.Name {
				t.Errorf("Name doesn't match")
			}
			if r.Lua != orig.Lua {
				t.Errorf("Lua doesn't match")
			}
		}
	})
}

func TestUpdateReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].reports); i++ {
			orig := data[0].reports[i]
			curr := d.reports[i]

			curr.Name = "blah"
			curr.Lua = "empty"

			r, err := updateReport(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error updating report: %s\n", err)
			}

			if r.ReportId != curr.ReportId {
				t.Errorf("ReportId doesn't match")
			}
			if r.Name != curr.Name {
				t.Errorf("Name doesn't match")
			}
			if r.Lua != curr.Lua {
				t.Errorf("Lua doesn't match")
			}
		}
	})
}

func TestDeleteReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 1; i < len(data[0].reports); i++ {
			orig := data[0].reports[i]
			curr := d.reports[i]

			err := deleteReport(d.clients[orig.UserId], &curr)
			if err != nil {
				t.Fatalf("Error deleting report: %s\n", err)
			}

			_, err = getReport(d.clients[orig.UserId], curr.ReportId)
			if err == nil {
				t.Fatalf("Expected error fetching deleted report")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error fetching deleted report: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error fetching deleted report")
			}
		}
	})
}
