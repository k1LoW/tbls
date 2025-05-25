package mariadb

import (
	"database/sql"

	"github.com/SouhlInc/tbls/drivers"
	"github.com/SouhlInc/tbls/drivers/mysql"
)

type Mariadb struct {
	mysql.Mysql
}

// New return new Mariadb.
func New(db *sql.DB, opts ...drivers.Option) (*Mariadb, error) {
	m, err := mysql.New(db, opts...)
	if err != nil {
		return nil, err
	}
	m.EnableMariaMode()
	return &Mariadb{*m}, nil
}
