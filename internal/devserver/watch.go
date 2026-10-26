package devserver

import (
	"context"
	"hash/fnv"
	"io/fs"
	"path/filepath"
	"strconv"
	"time"
)

const maxWatchedFiles = 20000

// Snapshot summarizes names, sizes and modification times under root. It
// polls rather than using OS notifications, so it behaves the same everywhere.
func Snapshot(root string) (uint64, error) {
	h := fnv.New64a()
	count := 0
	err := filepath.WalkDir(root, func(name string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if name != root && (d.Name() == ".git" || d.Name() == ".koru" || d.Name() == "node_modules") {
				return filepath.SkipDir
			}
			// Directory times change when ignored children change.
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if count++; count > maxWatchedFiles {
			return filepath.SkipAll
		}
		h.Write([]byte(name))
		h.Write([]byte(strconv.FormatInt(info.Size(), 10)))
		h.Write([]byte(strconv.FormatInt(info.ModTime().UnixNano(), 10)))
		return nil
	})
	return h.Sum64(), err
}

// Watch calls changed whenever a snapshot of any root differs from the last,
// and failed when a later snapshot cannot be taken.
func Watch(ctx context.Context, roots []string, every time.Duration, changed func(), failed func(error)) error {
	take := func() (uint64, error) {
		var sum uint64
		for _, root := range roots {
			s, err := Snapshot(root)
			if err != nil {
				return 0, err
			}
			sum = sum*31 + s
		}
		return sum, nil
	}
	last, err := take()
	if err != nil {
		return err
	}
	tick := time.NewTicker(every)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
			now, err := take()
			if err != nil {
				// A file can vanish mid-walk while an editor saves; report it
				// and try again on the next tick.
				failed(err)
				continue
			}
			if now != last {
				last = now
				changed()
			}
		}
	}
}
