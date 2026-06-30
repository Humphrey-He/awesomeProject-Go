package memory_model_practices

import "sync/atomic"

type Config struct {
	Version   uint64
	Threshold int
	Enabled   bool
}

type Publisher struct {
	ptr atomic.Pointer[Config]
}

func (p *Publisher) Store(cfg Config) {
	c := cfg
	p.ptr.Store(&c)
}

func (p *Publisher) Load() (Config, bool) {
	v := p.ptr.Load()
	if v == nil {
		return Config{}, false
	}
	return *v, true
}

type OnceFlag struct {
	done atomic.Bool
}

func (f *OnceFlag) Do(fn func()) bool {
	if f.done.CompareAndSwap(false, true) {
		fn()
		return true
	}
	return false
}
