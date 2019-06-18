package main

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/prometheus/common/log"
	"strings"
	"time"
)

// SQLServerConnection maps a connection to an SQL Server DB
type SQLServerConnection struct {
	connectionString string
	conn             *sql.DB
}

// StatData values of the statistics data
type StatData struct {
	ServerID string
	Value    string
	Date     time.Time
	StatType string
	Unit     string
	StatName string
	ItemName string
}

// Connect connect to the DB
func (connection *SQLServerConnection) Connect() error {
	conn, err := sql.Open("mssql", connection.connectionString)
	if err != nil {
		return err
	}
	connection.conn = conn

	return nil
}

// Close closes the DB connection
func (connection *SQLServerConnection) Close() error {
	return connection.conn.Close()
}

// GetStatData returns the stats for the servers id passed
func (connection *SQLServerConnection) GetStatData(serversID []string) ([]StatData, error) {
	inClause := strings.Join(serversID, ",")
	rows, err := connection.conn.Query(fmt.Sprintf("SELECT CompID, Value, Date, StatType, Unit, StatName, ItemName FROM StatData, Statistic WHERE StatData.StatID = Statistic.StatID and Statistic.CompID in (%s)", inClause))
	if err != nil {
		log.Fatal("DB Query failed:", err)
		return nil, err
	}
	defer rows.Close()
	var stats []StatData
	for rows.Next() {
		s := StatData{}
		if err := rows.Scan(&s.ServerID, &s.Value, &s.Date, &s.StatType, &s.Unit, &s.StatName, &s.ItemName); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if rows.Err() != nil {
		log.Fatal("failed to read all posts: " + rows.Err().Error())
		return nil, fmt.Errorf("failed to read all stats %s", rows.Err().Error())
	}
	return stats, nil
}
