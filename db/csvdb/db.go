package csvdb

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
)

type DB struct {
	accs *table
	ctr  *table
}

func Init(accounts, ctr string) (*DB, error) {
	d := &DB{
		accs: new(table),
		ctr:  new(table),
	}
	open := func(path string) (*os.File, error) {
		return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0644)
	}

	var err error
	if d.accs.File, err = open(accounts); err != nil {
		return nil, err
	}
	if d.ctr.File, err = open(ctr); err != nil {
		d.accs.Close()
		return nil, err
	}

	err = func() error {
		fi, err := d.ctr.Stat()
		if err != nil {
			return err
		}
		if fi.Size() == 0 {
			if _, err := fmt.Fprint(d.ctr, "0"); err != nil {
				return err
			}
		}
		return nil
	}()
	if err != nil {
		d.accs.Close()
		d.ctr.Close()
		return nil, err
	}

	return d, nil
}

func (d *DB) Close() error {
	var err error
	for _, c := range []io.Closer{d.accs, d.ctr} {
		if innerErr := c.Close(); innerErr != nil {
			err = innerErr
		}
	}
	return err
}

type table struct {
	sync.Mutex
	*os.File
}

func (t *table) all(ctx context.Context) ([][]string, error) {
	var recs [][]string

	return recs, withCtx(ctx, func() error {
		t.Lock()
		defer t.Unlock()
		var err error
		recs, err = t.allUnsync()
		return err
	})
}

func (t *table) allUnsync() ([][]string, error) {
	if _, err := t.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return csv.NewReader(t).ReadAll()
}

func (t *table) update(ctx context.Context, f func([][]string) ([][]string, error)) error {
	return withCtx(ctx, func() error {
		t.Lock()
		defer t.Unlock()
		records, err := t.allUnsync()
		if err != nil {
			return err
		}
		records, err = f(records)
		if err != nil {
			return err
		}
		if err := t.Truncate(0); err != nil {
			return err
		}
		if _, err := t.Seek(0, io.SeekStart); err != nil {
			return err
		}
		return csv.NewWriter(t).WriteAll(records)
	})
}

func (t *table) insert(ctx context.Context, record []string) error {
	return withCtx(ctx, func() error {
		t.Lock()
		defer t.Unlock()
		if _, err := t.Seek(0, io.SeekEnd); err != nil {
			return err
		}
		w := csv.NewWriter(t)
		if err := w.Write(record); err != nil {
			return err
		}
		w.Flush()
		return w.Error()
	})
}

func bumpCtr(ctx context.Context, t *table) (id int, err error) {
	err = withCtx(ctx, func() error {
		t.Lock()
		defer t.Unlock()

		if _, err := t.Seek(0, io.SeekStart); err != nil {
			return err
		}
		if _, err := fmt.Fscanf(t, "%d", &id); err != nil {
			return err
		}

		if err := t.Truncate(0); err != nil {
			return err
		}
		if _, err := t.Seek(0, io.SeekStart); err != nil {
			return err
		}
		_, err := fmt.Fprintf(t, "%d", id+1)
		return err
	})

	if err != nil {
		return -1, err
	}
	return id, nil
}

func withCtx(ctx context.Context, f func() error) error {
	c := make(chan error)
	go func() {
		select {
		case c <- f():
		default:
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-c:
		return err
	}
}
