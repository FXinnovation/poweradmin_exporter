package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/prometheus/common/log"
)

// SQLServerConnection maps a connection to an SQL Server DB
type SQLServerConnection struct {
	connectionString string
	conn             *sql.DB
}

// SQLServerInterface interface for sql server connection and methods
type SQLServerInterface interface {
	connect() error
	Close() error
	GetConnection() error
	GetStatData(serversID []string) ([]StatData, error)
	GetConfigComputerInfo(alias string) (ConfigComputerInfo, error)
	GetAllServersFor(group string) ([]string, error)
	GetAllServerMetric(serverID int) ([]ServerMetric, error)
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

// ConfigComputerInfo attributes of a server
type ConfigComputerInfo struct {
	CompID  int
	Name    string
	Alias   string
	GroupID int
}

// ServerMetric metric of a server
type ServerMetric struct {
	Name           string
	Alias          string
	GroupID        int
	StatID         int
	Value          float64
	Date           time.Time
	OwnerType      int
	ItemName       string
	StatName       string
	OwningComputer string
	CompID         int
	StatType       int
	ItemAlias      string
	Unit           int
	UnitStr        string
}

// NewSQLServerConnection returns a new connection object
func NewSQLServerConnection(connectionString string) *SQLServerConnection {
	sql := SQLServerConnection{
		connectionString: connectionString,
	}
	return &sql
}

// Connect to the DB
func (connection *SQLServerConnection) connect() error {
	conn, err := sql.Open("mssql", connection.connectionString)
	if err != nil {
		return err
	}
	connection.conn = conn
	return nil
}

// Close the DB connection
func (connection *SQLServerConnection) Close() error {
	return connection.conn.Close()
}

// GetConnection gets the connection to the DB
func (connection *SQLServerConnection) GetConnection() error {
	if connection.conn == nil {
		err := connection.connect()
		if err != nil {
			return err
		}
	}
	return nil
}

// GetStatData returns the stats for the given servers
func (connection *SQLServerConnection) GetStatData(serversID []string) ([]StatData, error) {
	inClause := strings.Join(serversID, ",")
	rows, err := connection.conn.Query(fmt.Sprintf("SELECT CompID, Value, Date, StatType, Unit, StatName, ItemName FROM StatData, Statistic WHERE StatData.StatID = Statistic.StatID and Statistic.CompID in (%s)", inClause))
	if err != nil {
		log.Error("DB Query failed:", err)
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
		log.Error("failed to read all posts: " + rows.Err().Error())
		return nil, fmt.Errorf("failed to read all stats %s", rows.Err().Error())
	}
	return stats, nil
}

// GetConfigComputerInfo get the attributes from a server
func (connection *SQLServerConnection) GetConfigComputerInfo(alias string) (ConfigComputerInfo, error) {
	sql := fmt.Sprintf(`
SELECT  TOP 1
		CompID,
		Name,
		Alias,
		GroupID
    FROM ConfigComputerInfo
  WHERE Alias = '%s'
`, alias)
	rows, err := connection.conn.Query(sql)
	if err != nil {
		log.Error("DB Query failed:", err)
		return ConfigComputerInfo{}, err
	}
	defer rows.Close()

	ci := ConfigComputerInfo{}
	for rows.Next() {
		if err := rows.Scan(&ci.CompID, &ci.Name, &ci.Alias, &ci.GroupID); err != nil {
			return ConfigComputerInfo{}, err
		}
	}
	if rows.Err() != nil {
		return ConfigComputerInfo{}, fmt.Errorf("error retrieving computer info: %s", rows.Err().Error())
	}
	return ci, nil
}

// GetAllServersFor get all the servers for a group
func (connection *SQLServerConnection) GetAllServersFor(group string) ([]string, error) {
	sql := fmt.Sprintf(`
SELECT CI.*
  FROM ConfigComputerInfo CI
	INNER JOIN (
		SELECT *
		  FROM ConfigGroupInfo
	) GI
	ON CI.GroupID = GI.GroupID
  WHERE GI.ParentGroupPath = '%s'
`, strings.ReplaceAll(group, "^", "\\"))

	log.Infof("query to be executed: %s\n", sql)

	rows, err := connection.conn.Query(sql)
	if err != nil {
		log.Error("DB Query failed:", err)
		return nil, err
	}
	defer rows.Close()
	var servList []string
	for rows.Next() {
		ci := ConfigComputerInfo{}
		if err := rows.Scan(&ci.CompID, &ci.Name, &ci.Alias, &ci.GroupID); err != nil {
			return nil, err
		}
		servList = append(servList, ci.Alias)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read all servers %s", rows.Err().Error())
	}
	return servList, nil
}

// GetAllServerMetric returns all server metric for a server
func (connection *SQLServerConnection) GetAllServerMetric(serverID int) ([]ServerMetric, error) {
	sql := fmt.Sprintf(`
SELECT 	CI.Name,
		CI.Alias,
		CI.GroupID,
		D.*
  FROM ConfigComputerInfo CI
      INNER JOIN (
          SELECT SD.*,
                 St.OwnerType,
                 St.ItemName,
                 St.StatName,
                 St.OwningComputer,
                 St.CompID,
                 St.StatType,
                 St.ItemAlias,
                 St.Unit,
                 St.UnitStr
             FROM
               (SELECT SD.StatID,
                       SD.Value,
                       SD.Date
                  FROM StatData SD
                      INNER JOIN (
                          SELECT StatID,
                                 MAX(Date) AS latest
                            FROM StatData
                            GROUP BY StatID
                      ) filteredSD ON filteredSD.StatID = SD.StatID
                                   AND filteredSD.latest = SD.Date
                ) SD
			  INNER JOIN Statistic St ON (SD.StatID = St.StatID)
			  WHERE St.ItemAlias <> ''
      ) D ON CI.CompID = D.CompID
  WHERE CI.CompID = %d
  ORDER BY D.StatID`, serverID)

	rows, err := connection.conn.Query(sql)
	if err != nil {
		log.Error("DB Query failed:", err)
		return nil, err
	}
	defer rows.Close()
	var metrics []ServerMetric
	for rows.Next() {
		m := ServerMetric{}
		if err := rows.Scan(
			&m.Name,
			&m.Alias,
			&m.GroupID,
			&m.StatID, &m.Value,
			&m.Date,
			&m.OwnerType,
			&m.ItemName,
			&m.StatName,
			&m.OwningComputer,
			&m.CompID,
			&m.StatType,
			&m.ItemAlias,
			&m.Unit,
			&m.UnitStr); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("failed to read all metrics %s", rows.Err().Error())
	}
	return metrics, nil
}
