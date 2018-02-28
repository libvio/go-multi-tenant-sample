package tenant

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// Db type
type Db struct {
	ID       int
	Name     string
	Host     string
	Port     int
	Dbname   string
	User     string
	Password string
}

var db *sqlx.DB
var catalog map[string]*sqlx.DB

// Init データベースに接続
func Init() error {
	catalog = make(map[string]*sqlx.DB)
	var err error
	db, err = sqlx.Open("mysql", "root:@(127.0.0.1:3306)/test_admin")
	return err
}

func GetTenantDb(name string) (*sqlx.DB, error) {
	connection, exist := catalog[name]
	if exist {
		return connection, nil
	}
	con := Db{}
	err := db.Get(&con, "SELECT * FROM db WHERE name = ? LIMIT 1", name)
	if err != nil {
		return nil, err
	}
	a, err := sqlx.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/%s", con.User, con.Password, con.Host, con.Port, con.Dbname))
	catalog[name] = a
	if err != nil {
		return nil, err
	}
	return a, nil
}
