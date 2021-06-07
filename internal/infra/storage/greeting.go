package storage

import (
	"context"
	"errors"
	"time"

	"github.com/bool64/brick-template/internal/domain/greeting"
	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/go-sql-driver/mysql"
)

// GreetingSaver saves greetings to database.
type GreetingSaver struct {
	Upstream greeting.Maker
	Storage  *sqluct.Storage
}

// GreetingsTable is the name of the table.
const GreetingsTable = "greetings"

// GreetingRow describes database mapping.
type GreetingRow struct {
	ID        int       `db:"id,omitempty"`
	Message   string    `db:"message"`
	CreatedAt time.Time `db:"created_at"`
}

// Hello makes a greeting with Upstream and stores it in database before returning.
func (gs *GreetingSaver) Hello(ctx context.Context, params greeting.Params) (string, error) {
	g, err := gs.Upstream.Hello(ctx, params)
	if err != nil {
		return g, err
	}

	q := gs.Storage.InsertStmt(GreetingsTable, GreetingRow{
		Message:   g,
		CreatedAt: time.Now(),
	})

	if _, err = gs.Storage.Exec(ctx, q); err != nil {
		var mySQLError *mysql.MySQLError

		if errors.As(err, &mySQLError) && mySQLError.Number == 1062 {
			// Duplicate entry error.
			return g, nil
		}

		return "", ctxd.WrapError(ctx, err, "failed to store greeting")
	}

	return g, nil
}

// GreetingMaker implements service provider.
func (gs *GreetingSaver) GreetingMaker() greeting.Maker {
	return gs
}
