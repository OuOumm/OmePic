package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

func (r *Repository) GetAllConfig(ctx context.Context) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT key, value FROM config`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, rows.Err()
}

func (r *Repository) UpsertConfigValues(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if _, err := r.db.ExecContext(
			ctx,
			`INSERT INTO config(key, value)
			 VALUES(?, ?)
			 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
			key,
			value,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) InsertMissingConfigValues(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if _, err := r.db.ExecContext(
			ctx,
			`INSERT INTO config(key, value)
			 VALUES(?, ?)
			 ON CONFLICT(key) DO NOTHING`,
			key,
			value,
		); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) GetConfigValue(ctx context.Context, key string) (string, error) {
	var value string
	err := r.db.QueryRowContext(ctx, `SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%w: config key %s", sql.ErrNoRows, key)
		}
		return "", err
	}
	return value, nil
}

func (r *Repository) SetConfigValue(ctx context.Context, key string, value string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value)
	return err
}
