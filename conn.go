package testdb

import (
	"database/sql/driver"
	"errors"
)

type conn struct {
	queries   map[string]*queries
	queryFunc func(query string, args []driver.Value) (driver.Rows, error)
	execFunc  func(query string, args []driver.Value) (driver.Result, error)
}

func newConn() *conn {
	return &conn{
		queries: make(map[string]*queries),
	}
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	s := new(stmt)

	if c.queryFunc != nil {
		s.queryFunc = func(args []driver.Value) (driver.Rows, error) {
			return c.queryFunc(query, args)
		}
	}

	if c.execFunc != nil {
		s.execFunc = func(args []driver.Value) (driver.Result, error) {
			return c.execFunc(query, args)
		}
	}

	if q, ok := d.conn.queries[getQueryHash(query)]; ok {
		if q.pos == len(q.queries) {
			return nil, errors.New("Exhausted stubs for query: " + query)
		}

		if s.queryFunc == nil && q.queries[q.pos].rows != nil {
			s.queryFunc = func(args []driver.Value) (driver.Rows, error) {
				defer func() {
					q.pos++
				}()

				if q.queries[q.pos].rows != nil {
					if rows, ok := q.queries[q.pos].rows.(*rows); ok {
						return rows.clone(), nil
					}
					return q.queries[q.pos].rows, nil
				}
				return nil, q.queries[q.pos].err
			}
		}

		if s.execFunc == nil && q.queries[q.pos].result != nil {
			s.execFunc = func(args []driver.Value) (driver.Result, error) {
				defer func() {
					q.pos++
				}()

				if q.queries[q.pos].result != nil {
					return q.queries[q.pos].result, nil
				}
				return nil, q.queries[q.pos].err
			}
		}
	}

	if !(s.queryFunc == nil || s.execFunc == nil) {
		return new(stmt), errors.New("Query not stubbed: " + query)
	}

	return s, nil
}

func (*conn) Close() error {
	return nil
}

func (*conn) Begin() (driver.Tx, error) {
	return &tx{}, nil
}

func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if c.queryFunc != nil {
		return c.queryFunc(query, args)
	}

	if q, ok := d.conn.queries[getQueryHash(query)]; ok {
		if q.pos == len(q.queries) {
			return nil, errors.New("Exhausted stubs for query: " + query)
		}

		defer func() {
			q.pos++
		}()

		return q.queries[q.pos].rows, q.queries[q.pos].err
	}

	return nil, errors.New("Query not stubbed: " + query)
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if c.execFunc != nil {
		return c.execFunc(query, args)
	}

	if q, ok := d.conn.queries[getQueryHash(query)]; ok {
		if q.pos == len(q.queries) {
			return nil, errors.New("Exhausted stubs for query: " + query)
		}

		defer func() {
			q.pos++
		}()

		if q.queries[q.pos].result != nil {
			return q.queries[q.pos].result, nil
		} else if q.queries[q.pos].err != nil {
			return nil, q.queries[q.pos].err
		}
	}

	return nil, errors.New("Exec call not stubbed: " + query)
}
