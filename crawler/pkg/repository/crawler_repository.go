package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/StevenRojas/natscrawler/crawler/config"
	"github.com/StevenRojas/natscrawler/crawler/pkg/model"
	_ "github.com/lib/pq"
	"time"
)

type Repository interface {
	Close() error
	AddURL(ctx context.Context, urlInfo model.UrlInfo) error
}

type repo struct {
	db *sql.DB
}

// NewRepository returns a new repository instance
func NewRepository(dbConf config.Database) (Repository, error) {
	conn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			dbConf.Host,
			dbConf.Username,
			dbConf.Password,
			dbConf.Database,
			dbConf.Port,
			dbConf.SSLMode,
		)
	db, err := sql.Open("postgres", conn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to database")
	return &repo{db: db}, nil
}

// Close closing DB connection
func (r *repo) Close() error {
	return r.db.Close()
}

// AddURL add URL info to the DB
func (r *repo) AddURL(ctx context.Context, ui model.UrlInfo) error {
	statement := `INSERT INTO public.ulrinfo(
	request_id, url, app_name, rating, rating_count, success, last_error, stats, created_at)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9);`

	stats, _ := json.Marshal(ui.Stats)
	stmt, err := r.db.PrepareContext(ctx, statement)
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, ui.RequestID, ui.Url, ui.AppName, ui.Rating, ui.RatingCount, ui.Success, ui.LastError, stats, time.Now().UTC())
	if err != nil {
		return err
	}
	fmt.Printf("App saved: %s\n", ui.AppName)
	return nil
}