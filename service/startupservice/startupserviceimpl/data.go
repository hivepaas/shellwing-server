package startupserviceimpl

import (
	"sync"
)

type startupData struct {
	mu sync.RWMutex
}

var (
	gStartupData = &startupData{
		mu: sync.RWMutex{},
	}
)

func (s *service) Shutdown() {
	if gStartupData == nil {
		panic("startup service shutdown")
	}

	gStartupData.mu.Lock()
	defer gStartupData.mu.Unlock()
	gStartupData = nil
}
