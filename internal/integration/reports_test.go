package integration_test

import (
	"github.com/aclindsa/moneygo/internal/handlers"
	"github.com/aclindsa/moneygo/internal/models"
	"net/http"
	"strconv"
	"testing"
)

func createReport(client *http.Client, report *models.Report) (*models.Report, error) {
	var r models.Report
	err := create(client, report, &r, "/v1/reports/")
	return &r, err
}

func getReport(client *http.Client, reportid int64) (*models.Report, error) {
	var r models.Report
	err := read(client, &r, "/v1/reports/"+strconv.FormatInt(reportid, 10))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func getReports(client *http.Client) (*models.ReportList, error) {
	var rl models.ReportList
	err := read(client, &rl, "/v1/reports/")
	if err != nil {
		return nil, err
	}
	return &rl, nil
}

func updateReport(client *http.Client, report *models.Report) (*models.Report, error) {
	var r models.Report
	err := update(client, report, &r, "/v1/reports/"+strconv.FormatInt(report.ReportId, 10))
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func deleteReport(client *http.Client, r *models.Report) error {
	err := remove(client, "/v1/reports/"+strconv.FormatInt(r.ReportId, 10))
	if err != nil {
		return err
	}
	return nil
}

func tabulateReport(client *http.Client, reportid int64) (*models.Tabulation, error) {
	var t models.Tabulation
	err := read(client, &t, "/v1/reports/"+strconv.FormatInt(reportid, 10)+"/tabulations")
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func TestCreateReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].reports); i++ {
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

			r.Lua = string(make([]byte, models.LuaMaxLength+1))
			_, err := createReport(d.clients[orig.UserId], &r)
			if err == nil {
				t.Fatalf("Expected error creating report with too-long Lua")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error creating report with too-long Lua: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error creating report with too-long Lua")
			}
		}
	})
}

func TestGetReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].reports); i++ {
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

func TestGetReports(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		rl, err := getReports(d.clients[0])
		if err != nil {
			t.Fatalf("Error fetching reports: %s\n", err)
		}

		numreports := 0
		foundIds := make(map[int64]bool)
		for i := 0; i < len(data[0].reports); i++ {
			orig := data[0].reports[i]
			curr := d.reports[i]

			if curr.UserId != d.users[0].UserId {
				continue
			}
			numreports += 1

			found := false
			for _, r := range *rl.Reports {
				if orig.Name == r.Name && orig.Lua == r.Lua {
					if _, ok := foundIds[r.ReportId]; ok {
						continue
					}
					foundIds[r.ReportId] = true
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Unable to find matching report: %+v", orig)
			}
		}

		if numreports != len(*rl.Reports) {
			t.Fatalf("Expected %d reports, received %d", numreports, len(*rl.Reports))
		}
	})
}

func TestUpdateReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].reports); i++ {
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

			r.Lua = string(make([]byte, models.LuaMaxLength+1))
			_, err = updateReport(d.clients[orig.UserId], r)
			if err == nil {
				t.Fatalf("Expected error updating report with too-long Lua")
			}
			if herr, ok := err.(*handlers.Error); ok {
				if herr.ErrorId != 3 { // Invalid requeset
					t.Fatalf("Unexpected API error updating report with too-long Lua: %s", herr)
				}
			} else {
				t.Fatalf("Unexpected error updating report with too-long Lua")
			}
		}
	})
}

func TestDeleteReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].reports); i++ {
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
func seriesEqualityHelper(t *testing.T, orig, curr map[string]*models.Series, name string) {
	if orig == nil || curr == nil {
		if orig != nil {
			t.Fatalf("`%s` series unexpectedly nil", name)
		}
		if curr != nil {
			t.Fatalf("`%s` series unexpectedly non-nil", name)
		}
		return
	}
	if len(orig) != len(curr) {
		t.Errorf("Series in question: %v\n", curr)
		t.Fatalf("Series' don't contain the same number of sub-series (found %d, expected %d)", len(curr), len(orig))
	}
	for k, os := range orig {
		cs := curr[k]
		if len(os.Values) != len(cs.Values) {
			t.Fatalf("`%s` series doesn't contain the same number of Values (found %d, expected %d)", k, len(cs.Values), len(os.Values))
		}
		for i, v := range os.Values {
			if v != cs.Values[i] {
				t.Errorf("Series doesn't contain the same values (found %f, expected %f)", cs.Values[i], v)
			}
		}
		seriesEqualityHelper(t, os.Series, cs.Series, k)
	}
}

func tabulationEqualityHelper(t *testing.T, orig, curr *models.Tabulation) {
	if orig.Title != curr.Title {
		t.Errorf("Tabulation Title doesn't match")
	}
	if orig.Subtitle != curr.Subtitle {
		t.Errorf("Tabulation Subtitle doesn't match")
	}
	if orig.Units != curr.Units {
		t.Errorf("Tabulation Units doesn't match")
	}
	if len(orig.Labels) != len(curr.Labels) {
		t.Fatalf("Tabulation doesn't contain the same number of labels")
	}
	for i, label := range orig.Labels {
		if label != curr.Labels[i] {
			t.Errorf("Label %d doesn't match", i)
		}
	}
	seriesEqualityHelper(t, orig.Series, curr.Series, "top-level")
}

func TestTabulateReport(t *testing.T) {
	RunWith(t, &data[0], func(t *testing.T, d *TestData) {
		for i := 0; i < len(data[0].tabulations); i++ {
			orig := data[0].tabulations[i]
			origReport := data[0].reports[orig.ReportId]
			report := d.reports[orig.ReportId]

			rt2, err := tabulateReport(d.clients[origReport.UserId], report.ReportId)
			if err != nil {
				t.Fatalf("Unexpected error tabulating report")
			}

			tabulationEqualityHelper(t, &orig, rt2)
		}
	})
}
