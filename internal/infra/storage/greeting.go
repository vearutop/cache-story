package storage

import (
	"context"
	"time"

	"github.com/bool64/ctxd"
	"github.com/bool64/sqluct"
	"github.com/bool64/stats"
	"github.com/vearutop/cache-story/internal/domain/greeting"
)

// GreetingSaver saves greetings to database.
type GreetingSaver struct {
	Upstream greeting.Maker
	Storage  *sqluct.Storage
	Stats    stats.Tracker
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
	}).Options("IGNORE")

	if _, err = gs.Storage.Exec(ctx, q); err != nil {
		return "", ctxd.WrapError(ctx, err, "failed to store greeting")
	}

	if err = gs.calcAvgID(ctx); err != nil {
		return "", err
	}

	return g, nil
}

// calcAvgID makes things sloooooower 🐌.
func (gs *GreetingSaver) calcAvgID(ctx context.Context) error {
	r := gs.Storage.DB().QueryRowContext(ctx, "SELECT AVG(id) FROM "+GreetingsTable+" WHERE id < 1000")

	var avg int

	if err := r.Scan(&avg); err != nil {
		return r.Err()
	}

	gs.Stats.Set(ctx, "avg_id", float64(avg))

	return nil
}

// ClearGreetings removes all entries.
func (gs *GreetingSaver) ClearGreetings(ctx context.Context) (int, error) {
	res, err := gs.Storage.DeleteStmt(GreetingsTable).ExecContext(ctx)
	if err != nil {
		return 0, err
	}

	aff, err := res.RowsAffected()
	return int(aff), err
}

// GreetingMaker implements service provider.
func (gs *GreetingSaver) GreetingMaker() greeting.Maker {
	return gs
}

// GreetingClearer implements service provider.
func (gs *GreetingSaver) GreetingClearer() greeting.Clearer {
	return gs
}
