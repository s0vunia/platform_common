package pg

import (
	"context"
	"fmt"
	"log"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/s0vunia/platform_common/pkg/db"
	"github.com/s0vunia/platform_common/pkg/db/prettier"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type key string

const (
	// TxKey is a constant key used to store and retrieve a transaction object in context.Context
	TxKey key = "tx"
)

type pg struct {
	dbc *pgxpool.Pool
}

// NewDB creates a new database connection using the provided pgxpool.Pool and returns a db.DB interface.
//
// Parameters:
// - pool: the pgxpool.Pool to use for the database connection
//
// Returns:
// - db.DB: the database connection
func NewDB(dbc *pgxpool.Pool) db.DB {
	return &pg{
		dbc: dbc,
	}
}

func (p *pg) ScanOneContext(ctx context.Context, dest interface{}, q db.Query, args ...interface{}) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, q.Name, opentracing.Tag{Key: "query", Value: q.QueryRaw})
	defer span.Finish()
	logQuery(ctx, q, args...)

	row, err := p.QueryContext(ctx, q, args...)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err.Error())
		return err
	}

	err = pgxscan.ScanOne(dest, row)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err.Error())
		return err
	}

	return nil
}

func (p *pg) ScanAllContext(ctx context.Context, dest interface{}, q db.Query, args ...interface{}) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, q.Name, opentracing.Tag{Key: "query", Value: q.QueryRaw})
	defer span.Finish()
	logQuery(ctx, q, args...)

	rows, err := p.QueryContext(ctx, q, args...)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err.Error())
		return err
	}

	err = pgxscan.ScanAll(dest, rows)
	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err.Error())
		return err
	}

	return nil
}

func (p *pg) ExecContext(ctx context.Context, q db.Query, args ...interface{}) (pgconn.CommandTag, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, q.Name, opentracing.Tag{Key: "query", Value: q.QueryRaw})
	defer span.Finish()
	logQuery(ctx, q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	var err error
	var pgc pgconn.CommandTag
	if ok {
		pgc, err = tx.Exec(ctx, q.QueryRaw, args...)
	} else {
		pgc, err = p.dbc.Exec(ctx, q.QueryRaw, args...)
	}

	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err.Error())
		return pgc, err
	}

	return pgc, nil
}

func (p *pg) QueryContext(ctx context.Context, q db.Query, args ...interface{}) (pgx.Rows, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, q.Name, opentracing.Tag{Key: "query", Value: q.QueryRaw})
	defer span.Finish()
	logQuery(ctx, q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	var rows pgx.Rows
	var err error
	if ok {
		rows, err = tx.Query(ctx, q.QueryRaw, args...)
	} else {
		rows, err = p.dbc.Query(ctx, q.QueryRaw, args...)
	}

	if err != nil {
		ext.Error.Set(span, true)
		span.SetTag("err", err.Error())
		return rows, err
	}

	return rows, nil
}

func (p *pg) QueryRowContext(ctx context.Context, q db.Query, args ...interface{}) pgx.Row {
	span, ctx := opentracing.StartSpanFromContext(ctx, q.Name, opentracing.Tag{Key: "query", Value: q.QueryRaw})
	defer span.Finish()
	logQuery(ctx, q, args...)

	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, q.QueryRaw, args...)
	}
	return p.dbc.QueryRow(ctx, q.QueryRaw, args...)
}

func (p *pg) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "BeginTx")
	defer span.Finish()
	return p.dbc.BeginTx(ctx, txOptions)
}

func (p *pg) Ping(ctx context.Context) error {
	return p.dbc.Ping(ctx)
}

func (p *pg) Close() {
	p.dbc.Close()
}

// MakeContextTx creates a new context with a transaction object attached.
//
// Parameters:
// - ctx: the context to attach the transaction object to
// - tx: the transaction object to attach
//
// Returns:
// - context.Context: the new context with the transaction object attached
func MakeContextTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}

func logQuery(ctx context.Context, q db.Query, args ...interface{}) {
	prettyQuery := prettier.Pretty(q.QueryRaw, prettier.PlaceholderDollar, args...)
	log.Println(
		ctx,
		fmt.Sprintf("sql: %s", q.Name),
		fmt.Sprintf("query: %s", prettyQuery),
	)
}
