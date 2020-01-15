package db

import (
	"encoding/csv"
	"io"
	"os"
	"sync"
	"time"
)

type NullTime struct {
	Time  time.Time
	Valid bool
}

type DB struct {
	sync.RWMutex
	*os.File
	ctr int
}

func Init(accounts string) (*DB, error) {
	d := new(DB)

	var err error
	d.File, err = os.OpenFile(accounts, os.O_RDWR|os.O_CREATE|os.O_SYNC, 0644)
	if err != nil {
		return nil, err
	}

	accs, err := d.Accounts()
	if err != nil {
		d.File.Close()
		return nil, err
	}
	for _, a := range accs {
		if a.ID > d.ctr {
			d.ctr = a.ID
		}
	}
	d.ctr++

	return d, nil
}

func (d *DB) all() ([][]string, error) {
	d.RLock()
	defer d.RUnlock()
	return d.allUnsync()
}

func (d *DB) allUnsync() ([][]string, error) {
	if _, err := d.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	return csv.NewReader(d).ReadAll()
}

func (d *DB) update(f func([][]string) ([][]string, error)) error {
	d.Lock()
	defer d.Unlock()

	records, err := d.allUnsync()
	if err != nil {
		return err
	}
	records, err = f(records)
	if err != nil {
		return err
	}

	if err := d.Truncate(0); err != nil {
		return err
	}
	if _, err := d.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return csv.NewWriter(d).WriteAll(records)
}
