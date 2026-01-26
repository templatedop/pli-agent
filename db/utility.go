package db

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// validateOutputVariable ensures the output variable is not nil
func validateOutputVariable(result interface{}) error {
	if result == nil {
		return pgx.ErrNoRows
	}
	return nil
}

// QueueExecRow queues an execution query in a batch (for INSERT/UPDATE/DELETE)
// Supports both Squirrel builder and raw SQL
func QueueExecRow(batch *pgx.Batch, builder sq.Sqlizer) error {
	var qErr error
	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	batch.Queue(sql, args...).Exec(func(ct pgconn.CommandTag) error {
		rowsAffected := ct.RowsAffected()
		if rowsAffected == 0 {
			qErr = pgx.ErrNoRows
			return nil
		}
		return nil
	})
	return qErr
}

// QueueExecRowRaw queues a raw SQL execution query in a batch
// Use this for CTE queries that Squirrel doesn't support
func QueueExecRowRaw(batch *pgx.Batch, sql string, args ...interface{}) error {
	var qErr error
	batch.Queue(sql, args...).Exec(func(ct pgconn.CommandTag) error {
		rowsAffected := ct.RowsAffected()
		if rowsAffected == 0 {
			qErr = pgx.ErrNoRows
			return nil
		}
		return nil
	})
	return qErr
}

// QueueReturn queues a query that returns multiple rows in a batch
// Supports both Squirrel builder and raw SQL
func QueueReturn[T any](batch *pgx.Batch, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *[]T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}
	var qErr error
	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		collectedRows, err := pgx.CollectRows(rows, scanFn)
		if err != nil {
			qErr = err
			return nil
		}
		*result = collectedRows
		return nil
	})
	return qErr
}

// QueueReturnRaw queues a raw SQL query that returns multiple rows in a batch
// Use this for CTE queries that Squirrel doesn't support
func QueueReturnRaw[T any](batch *pgx.Batch, sql string, args []interface{}, scanFn pgx.RowToFunc[T], result *[]T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}
	var qErr error
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		collectedRows, err := pgx.CollectRows(rows, scanFn)
		if err != nil {
			qErr = err
			return nil
		}
		*result = collectedRows
		return nil
	})
	return qErr
}

// QueueReturnRow queues a query that returns a single row in a batch
// Supports both Squirrel builder and raw SQL
func QueueReturnRow[T any](batch *pgx.Batch, builder sq.Sqlizer, scanFn pgx.RowToFunc[T], result *T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}
	var qErr error
	sql, args, err := builder.ToSql()
	if err != nil {
		return err
	}
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		collectedRow, err := pgx.CollectOneRow(rows, scanFn)
		if err != nil {
			qErr = err
			return nil
		}
		*result = collectedRow
		return nil
	})
	return qErr
}

// QueueReturnRowRaw queues a raw SQL query that returns a single row in a batch
// Use this for CTE queries that Squirrel doesn't support
func QueueReturnRowRaw[T any](batch *pgx.Batch, sql string, args []interface{}, scanFn pgx.RowToFunc[T], result *T) error {
	if err := validateOutputVariable(result); err != nil {
		return err
	}
	var qErr error
	batch.Queue(sql, args...).Query(func(rows pgx.Rows) error {
		collectedRow, err := pgx.CollectOneRow(rows, scanFn)
		if err != nil {
			qErr = err
			return nil
		}
		*result = collectedRow
		return nil
	})
	return qErr
}
