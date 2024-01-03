package passphrase

import "sync"

type Passphrase struct {
	sync.RWMutex
	val []byte
}

func New() *Passphrase {
	return &Passphrase{}
}

func (p *Passphrase) Set(val []byte) {
	p.Lock()
	defer p.Unlock()

	p.val = val
}

func (p *Passphrase) Get() string {
	p.RLock()
	defer p.RUnlock()

	//TODO: clean up val with crypto/subtle
	return string(p.val)
}
