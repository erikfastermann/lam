package csvdb

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sync"
)

// TODO: use context
type DB struct {
	accs  *table
	users *table
	ctr   *table
}

func Init(users, accounts, ctr string) (*DB, error) {
	d := &DB{
		accs:  new(table),
		users: new(table),
		ctr:   new(table),
	}
	open := func(path string) (*os.File, error) {
		return os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0644)
	}

	var err error
	if d.users.File, err = open(users); err != nil {
		return nil, err
	}
	if d.accs.File, err = open(accounts); err != nil {
		d.users.Close()
		return nil, err
	}
	if d.ctr.File, err = open(ctr); err != nil {
		d.users.Close()
		d.accs.Close()
		return nil, err
	}

	fi, err := d.ctr.Stat()
	if err != nil {
		d.users.Close()
		d.accs.Close()
		d.ctr.Close()
		return nil, err
	}
	if fi.Size() == 0 {
		if _, err := fmt.Fprintf(d.ctr, "%d,%d", 0, 0); err != nil {
			d.users.Close()
			d.accs.Close()
			d.ctr.Close()
			return nil, err
		}
	}

	return d, nil
}

func (d *DB) Close() error {
	var err error
	for _, c := range []io.Closer{d.users, d.accs, d.ctr} {
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

func (t *table) all() ([][]string, error) {
	t.Lock()
	defer t.Unlock()
	return t.allUnsync()
}

func (t *table) allUnsync() ([][]string, error) {
	if _, err := t.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return csv.NewReader(t).ReadAll()
}

func (t *table) update(f func([][]string) ([][]string, error)) error {
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
}

func (t *table) insert(record []string) error {
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
}

const (
	ctrPosUser = 0
	ctrPosAcc  = 1
)

func bumpCtr(t *table, ctrPos uint) (int, error) {
	bump1, bump2 := 0, 0
	switch ctrPos {
	case 0:
		bump1 = 1
	case 1:
		bump2 = 1
	default:
		panic("invalid counter")
	}
	t.Lock()
	defer t.Unlock()

	if _, err := t.Seek(0, io.SeekStart); err != nil {
		return -1, err
	}
	var id1, id2 int
	if _, err := fmt.Fscanf(t, "%d,%d", &id1, &id2); err != nil {
		return -1, err
	}

	if err := t.Truncate(0); err != nil {
		return -1, err
	}
	if _, err := t.Seek(0, io.SeekStart); err != nil {
		return -1, err
	}
	if _, err := fmt.Fprintf(t, "%d,%d", id1+bump1, id2+bump2); err != nil {
		return -1, err
	}

	if bump1 > 0 {
		return id1, nil
	}
	return id2, nil
}