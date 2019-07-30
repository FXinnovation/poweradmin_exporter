package main

import (
	"database/sql/driver"
	"github.com/DATA-DOG/go-sqlmock"
	"reflect"
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

func TestSQLServerConnection_GetConfigComputerInfo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	connection := &SQLServerConnection{
		conn: db,
	}
	rows := sqlmock.NewRows([]string{"CompID", "Name", "Alias", "GroupID"}).
		AddRow(2, "FXSERVER", "FXSERVER", 1)
	mock.ExpectQuery("SELECT (.+) FROM ConfigComputerInfo WHERE Alias = (.+)$").WillReturnRows(rows)
	ci, err := connection.GetConfigComputerInfo("FXSERVER")
	if err != nil {
		t.Fatalf("Get CI returned an error:%v", err)
	}
	if ci.CompID != 2 {
		t.Fatalf("Get CI returned bad data")
	}
	// we make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestSQLServerConnection_GetAllServersFor(t *testing.T) {

	type sqlReturn struct {
		CompID  int
		Name    string
		Alias   string
		GroupID int
	}
	type fields struct {
		connection *SQLServerConnection
		rows       []sqlReturn
	}
	type args struct {
		group string
	}
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	connection := &SQLServerConnection{
		conn: db,
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "gotresult",
			fields: fields{
				connection: connection, rows: []sqlReturn{{CompID: 2, Name: "FXSERVER", Alias: "FXSERVER", GroupID: 1}},
			},
			args: args{
				group: "Servers/Devices^Live^Dev",
			},
			want: []string{
				"FXSERVER",
			},
			wantErr: false,
		},
		{
			name: "noresult",
			fields: fields{
				connection: connection, rows: []sqlReturn{},
			},
			args: args{
				group: "NoReturn",
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := sqlmock.NewRows([]string{"CompID", "Name", "Alias", "GroupID"})
			for _, r := range tt.fields.rows {
				rows.AddRow(r.CompID, r.Name, r.Alias, r.GroupID)
			}
			const expectedSQL = "SELECT (.*) FROM ConfigComputerInfo CI INNER JOIN (.+) GI ON CI.GroupID = GI.GroupID WHERE GI.ParentGroupPath = (.+)"
			mock.ExpectQuery(expectedSQL).WillReturnRows(rows)
			got, err := tt.fields.connection.GetAllServersFor(tt.args.group)
			// we make sure that all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAllServersFor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetAllServersFor() got = %v, want %v", got, tt.want)
			}
		})
	}
}
