package autoid

import (
	"strconv"
	"sync"

	"github.com/AlbinoDrought/creamy-videos-importer/creamqueue"
)

// AutoID is a goroutine-safe way to generate unique IDs
type AutoID interface {
	Next() creamqueue.JobID
}

type locky struct {
	id   uint64
	lock sync.Locker
}

func (repo *locky) Next() creamqueue.JobID {
	repo.lock.Lock()
	defer repo.lock.Unlock()

	id := strconv.FormatUint(repo.id, 16)
	repo.id++
	return creamqueue.JobID(id)
}

// Make a locking goroutine-safe unique ID generator
func Make() AutoID {
	return &locky{
		lock: &sync.Mutex{},
	}
}
