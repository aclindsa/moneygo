package handlers

import (
	"github.com/aclindsa/moneygo/internal/models"
	"github.com/aclindsa/moneygo/internal/reports"
	"github.com/aclindsa/moneygo/internal/store"
	"log"
	"net/http"
)

func ReportTabulationHandler(tx store.Tx, r *http.Request, user *models.User, reportid int64) ResponseWriterWriter {
	report, err := tx.GetReport(reportid, user.UserId)
	if err != nil {
		return NewError(3 /*Invalid Request*/)
	}

	tabulation, err := reports.RunReport(tx, user, report)
	if err != nil {
		// TODO handle different failure cases differently
		log.Print("reports.RunReport returned:", err)
		return NewError(3 /*Invalid Request*/)
	}

	tabulation.ReportId = reportid

	return tabulation
}

func ReportHandler(r *http.Request, context *Context) ResponseWriterWriter {
	user, err := GetUserFromSession(context.Tx, r)
	if err != nil {
		return NewError(1 /*Not Signed In*/)
	}

	if r.Method == "POST" {
		var report models.Report
		if err := ReadJSON(r, &report); err != nil {
			return NewError(3 /*Invalid Request*/)
		}
		report.ReportId = -1
		report.UserId = user.UserId

		if len(report.Lua) >= models.LuaMaxLength {
			return NewError(3 /*Invalid Request*/)
		}

		err = context.Tx.InsertReport(&report)
		if err != nil {
			log.Print(err)
			return NewError(999 /*Internal Error*/)
		}

		return ResponseWrapper{201, &report}
	} else if r.Method == "GET" {
		if context.LastLevel() {
			//Return all Reports
			var rl models.ReportList
			reports, err := context.Tx.GetReports(user.UserId)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}
			rl.Reports = reports
			return &rl
		}

		reportid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		if context.NextLevel() == "tabulations" {
			return ReportTabulationHandler(context.Tx, r, user, reportid)
		} else {
			// Return Report with this Id
			report, err := context.Tx.GetReport(reportid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			return report
		}
	} else {
		reportid, err := context.NextID()
		if err != nil {
			return NewError(3 /*Invalid Request*/)
		}

		if r.Method == "PUT" {
			var report models.Report
			if err := ReadJSON(r, &report); err != nil || report.ReportId != reportid {
				return NewError(3 /*Invalid Request*/)
			}
			report.UserId = user.UserId

			if len(report.Lua) >= models.LuaMaxLength {
				return NewError(3 /*Invalid Request*/)
			}

			err = context.Tx.UpdateReport(&report)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return &report
		} else if r.Method == "DELETE" {
			report, err := context.Tx.GetReport(reportid, user.UserId)
			if err != nil {
				return NewError(3 /*Invalid Request*/)
			}

			err = context.Tx.DeleteReport(report)
			if err != nil {
				log.Print(err)
				return NewError(999 /*Internal Error*/)
			}

			return SuccessWriter{}
		}
	}
	return NewError(3 /*Invalid Request*/)
}
