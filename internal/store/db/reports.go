package db

import (
	"fmt"
	"github.com/aclindsa/moneygo/internal/models"
)

func (tx *Tx) GetReport(reportid int64, userid int64) (*models.Report, error) {
	var r models.Report

	err := tx.SelectOne(&r, "SELECT * from reports where UserId=? AND ReportId=?", userid, reportid)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (tx *Tx) GetReports(userid int64) (*[]*models.Report, error) {
	var reports []*models.Report

	_, err := tx.Select(&reports, "SELECT * from reports where UserId=?", userid)
	if err != nil {
		return nil, err
	}
	return &reports, nil
}

func (tx *Tx) InsertReport(report *models.Report) error {
	err := tx.Insert(report)
	if err != nil {
		return err
	}
	return nil
}

func (tx *Tx) UpdateReport(report *models.Report) error {
	count, err := tx.Update(report)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to update 1 report, was going to update %d", count)
	}
	return nil
}

func (tx *Tx) DeleteReport(report *models.Report) error {
	count, err := tx.Delete(report)
	if err != nil {
		return err
	}
	if count != 1 {
		return fmt.Errorf("Expected to delete 1 report, was going to delete %d", count)
	}
	return nil
}
