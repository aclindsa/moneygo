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
