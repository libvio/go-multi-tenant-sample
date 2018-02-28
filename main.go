package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"

	"./tenant"

	_ "github.com/go-sql-driver/mysql"
)

// Item type
type Item struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

func main() {
	err := tenant.Init()
	if err != nil {
		panic(err.Error())
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/{tenantName:[a-zA-Z0-9-]+}", func(r chi.Router) {
		r.Use(tenantCtx)
		r.Get("/item", getItem)
	})
	http.ListenAndServe(":3333", r)
}

type contextKey string

const tenantContextKey contextKey = "tenantDb"

func tenantCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantName := chi.URLParam(r, "tenantName")
		tenantDb, err := tenant.GetTenantDb(tenantName)
		if err != nil {
			panic(err.Error())
		}
		ctx := context.WithValue(r.Context(), tenantContextKey, tenantDb)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getItem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantDb := ctx.Value(tenantContextKey).(*sqlx.DB)
	item := Item{}
	err := tenantDb.Get(&item, "SELECT * FROM item LIMIT 1")
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
	}
	jsonBytes, _ := json.Marshal(item)
	w.Write(jsonBytes)
}
