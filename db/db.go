package db

import (
	"encoding/csv"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type NullTime struct {
	Time  time.Time
	Valid bool
}

type DB struct {
	dir string
	sync.RWMutex
	*os.File
	ctr int
}

const accFile = "accounts.csv"

func Init(dir string) (*DB, error) {
	d := &DB{dir: dir}

	if err := d.open(); err != nil {
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

func (d *DB) open() error {
	f, err := os.OpenFile(
		filepath.Join(d.dir, accFile),
		os.O_RDWR|os.O_CREATE|os.O_SYNC,
		0644,
	)
	if err != nil {
		return err
	}
	d.File = f
	return nil
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

	tmp, err := ioutil.TempFile(d.dir, accFile)
	if err != nil {
		return err
	}
	name := tmp.Name()

	wErr := csv.NewWriter(tmp).WriteAll(records)
	if err := tmp.Close(); err != nil {
		return err
	}
	if wErr != nil {
		return err
	}

	if err := os.Rename(name, filepath.Join(d.dir, accFile)); err != nil {
		return err
	}
	return d.open()
}
