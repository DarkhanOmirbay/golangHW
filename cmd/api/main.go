package main

import (
	"context"
	"database/sql"
	"flag"
	_ "github.com/lib/pq"
	"golangHW.darkhanomirbay/internal/data"
	"golangHW.darkhanomirbay/internal/jsonlog"
	"golangHW.darkhanomirbay/internal/mailer"
	"os"
	"sync"
	"time"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 5000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://postgres:703905@localhost/d.omirbayDB?sslmode=disable", "PostgreSQL DSN")

	flag.IntVar(&cfg.db.maxOpenConns, "max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "29e20c93d498ff", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "4672af6936d913", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "<220373@astanait.edu.kz> ", "SMTP sender")
	flag.Parse()

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("database connection pool established", nil)
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	app.gorout()

	err = app.serve()
	if err != nil {
		app.logger.PrintFatal(err, nil)
	}

}
func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (app *application) gorout() {
	for {
		//var filters data.Filters
		//users, _, err := app.models.UserInfoModel.GetAllNonActivated("", " ", filters)
		users, err := app.models.UserInfoModel.GetAllNoActiv()
		if err != nil {
			app.logger.PrintError(err, nil)

		}

		for _, user := range users {
			err := app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
			if err != nil {
				app.logger.PrintError(err, nil)
			}
			token, err := app.models.Tokens.New(user.ID, 2*time.Minute, data.ScopeActivation)
			if err != nil {
				app.logger.PrintError(err, nil)
				return
			}
			go func() {
				// As there are now multiple pieces of data that we want to pass to our email
				// templates, we create a map to act as a 'holding structure' for the data. This
				// contains the plaintext version of the activation token for the user, along
				// with their ID.
				data := map[string]any{
					"activationToken": token.Plaintext,
					"userID":          user.ID,
				}
				err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
				if err != nil {
					app.logger.PrintError(err, nil)
				}
			}()
		}
	}
}
