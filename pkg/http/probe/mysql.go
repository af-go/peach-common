package probe

import (
	"database/sql"
	"fmt"
	"time"

	model "github.com/af-go/peach-common/pkg/model/database"
	"github.com/go-logr/logr"
	_ "github.com/go-sql-driver/mysql"
)

func BuildMySQLProbe(options model.MySQLOptions, logger *logr.Logger) *MySQLProbe {
	c := MySQLProbe{options: options, logger: logger}
	err := c.init()
	if err != nil {
		logger.Error(err, "failed to init mysql/mariadb probe")
		return nil
	}
	return &c
}

type MySQLProbe struct {
	db      *sql.DB
	options model.MySQLOptions
	logger  *logr.Logger
}

func (c *MySQLProbe) init() error {
	var err error
	dsn := fmt.Sprintf("%s:%s@(%s:%d)/%s", c.options.Username, c.options.Password, c.options.Host, c.options.Port, c.options.Database)
	c.db, err = sql.Open("mysql", dsn)
	if err != nil {
		c.logger.Error(err, "failed to open mysql/mariadb connection", "dsn", dsn)
		return err
	}
	c.db.SetConnMaxLifetime(time.Minute * 3)
	c.db.SetMaxOpenConns(10)
	c.db.SetMaxIdleConns(10)
	return nil
}

func (c *MySQLProbe) Do() bool {
	stmt, err := c.db.Prepare("SELECT 1")
	if err != nil {
		c.logger.Error(err, "failed create statement")
		return false
	}
	_, err = stmt.Query()
	if err != nil {
		c.logger.Error(err, "failed to execuate query")
		return false
	}
	return true
}
