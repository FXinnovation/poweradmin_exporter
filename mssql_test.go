package main

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"testing"
	"time"
)

func TestSqlServerConnection_GetStatData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	connection := &SQLServerConnection{
		conn: db,
	}
	rowTime, _ := time.Parse("2006-01-02 15:04:05.000", "2019-04-24 20:41:35.000")
	rows := sqlmock.NewRows([]string{"CompID", "Value", "Date", "StatType", "Unit", "StatName", "ItemName"}).
		AddRow(2, 118180806656, driver.Value(rowTime), 22, 1, "Free Bytes", "C:")
	mock.ExpectQuery("^SELECT (.+) FROM StatData, Statistic WHERE StatData.StatID = Statistic.StatID and Statistic.CompID in (.+)$").WillReturnRows(rows)
	serversID := []string{"2", "3"}
	stats, err := connection.GetStatData(serversID)
	if err != nil {
		t.Fatalf("Get stats ddata returned an error:%v", err)
	}
	if len(stats) == 0 {
		t.Fatalf("Get stats data returned no results")
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
