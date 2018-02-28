package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"unsafe"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	_ "github.com/go-sql-driver/mysql"
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

type Item struct {
	Code string
	Name string
}

func main() {
	tenantDbs := getTenantList()
	fmt.Println(tenantDbs)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/{tenantName:[a-zA-Z0-9-]+}", func(r chi.Router) {
		r.Use(TenantCtx)
		r.Get("/item", getItem)
	})
	http.ListenAndServe(":3333", r)
}

type contextKey string

const tokenContextKey contextKey = "tenantDb"

func TenantCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantName := chi.URLParam(r, "tenantName")
		tenantDb := getTenant(getTenantList(), tenantName)
		ctx := context.WithValue(r.Context(), tokenContextKey, tenantDb)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantDb := ctx.Value(tokenContextKey).(Db)
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@(%s:%d)/%s", tenantDb.User, tenantDb.Password, tenantDb.Host, tenantDb.Port, tenantDb.Dbname))
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM item")
	defer rows.Close()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	items := []Item{}
	item := Item{}
	for rows.Next() {
		err = rows.Scan(&item.Code, &item.Name)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		items = append(items, item)
	}
	w.Write(sbytes(items[0].Name))
	return
}

func sbytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func getTenantList() []Db {
	db, err := sql.Open("mysql", "root:@(127.0.0.1:3306)/test_admin")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM db")
	defer rows.Close()
	if err != nil {
		panic(err.Error())
	}
	tenantDbs := []Db{}
	tenantDb := Db{}
	for rows.Next() {
		err = rows.Scan(&tenantDb.ID, &tenantDb.Name, &tenantDb.Host, &tenantDb.Port, &tenantDb.Dbname, &tenantDb.User, &tenantDb.Password)
		if err != nil {
			panic(err.Error())
		}
		tenantDbs = append(tenantDbs, tenantDb)
	}
	return tenantDbs
}

func getTenant(tenantDbs []Db, name string) Db {
	for _, v := range tenantDbs {
		if v.Name == name {
			return v
		}
	}
	return Db{}
}
