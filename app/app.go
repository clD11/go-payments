package app

import (
	"database/sql"
	"fmt"
	"github.com/clD11/form3-payments/handler"
	"github.com/clD11/form3-payments/model"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"time"
)

type App struct {
	Router *mux.Router
	DB     *pg.DB
}

func (a *App) Initialize(config *Config) {
	a.createDatabaseAndMigration(config)
	a.registerRoutes()
}

func (a *App) GetPayment(w http.ResponseWriter, r *http.Request) {
	handler.GetPayment(a.DB, w, r)
}

func (a *App) CreatePayment(w http.ResponseWriter, r *http.Request) {
	handler.CreatePayment(a.DB, w, r)
}

func (a *App) DeletePayment(w http.ResponseWriter, r *http.Request) {
	handler.DeletePayment(a.DB, w, r)
}

func (a *App) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	handler.UpdatePayment(a.DB, w, r)
}

func (a *App) GetPayments(w http.ResponseWriter, r *http.Request) {
	handler.GetPayments(a.DB, w, r)
}

func (a *App) Run(host string) {
	log.Fatal(http.ListenAndServe(host, a.Router))
}

func (a *App) createDatabaseAndMigration(config *Config) {
	a.ping(config)

	db := pg.Connect(config.DB)

	tables := []interface{}{
		(*model.Payment)(nil),
		(*model.Attributes)(nil),
		(*model.BeneficiaryParty)(nil),
		(*model.ChargesInformation)(nil),
		(*model.SponsorParty)(nil),
		(*model.DebtorParty)(nil),
		(*model.Charge)(nil),
		(*model.Fx)(nil)}

	for _, model := range tables {
		db.Options()
		err := db.CreateTable(model, &orm.CreateTableOptions{})
		if err != nil {
			log.Print(err)
		}
	}

	a.DB = db
}

func (a *App) ping(config *Config) {
	conn, err := sql.Open("postgres",
		fmt.Sprintf("host=postgres port=5432 user=%s password=%s dbname=%d sslmode=disable",
			config.DB.User, config.DB.Password, config.DB.Database))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// ping the database until it comes up
	timeout := time.Now().Add(time.Second * 20)
	for time.Now().Before(timeout) {
		if err = conn.Ping(); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (a *App) registerRoutes() {
	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/v1/payments/{id}", a.GetPayment).Methods(http.MethodGet)
	a.Router.HandleFunc("/v1/payments", a.CreatePayment).Methods(http.MethodPost)
	a.Router.HandleFunc("/v1/payments/{id}", a.DeletePayment).Methods(http.MethodDelete)
	a.Router.HandleFunc("/v1/payments/{id}", a.UpdatePayment).Methods(http.MethodPut)
	a.Router.HandleFunc("/v1/payments", a.GetPayments).Methods(http.MethodGet)
}
