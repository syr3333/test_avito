package repository

import (
	"context"
	"database/sql"
)

type txManager struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) TransactionManager {
	return &txManager{db: db}
}

func (tm *txManager) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return tm.db.BeginTx(ctx, nil)
}
