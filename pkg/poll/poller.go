package poll

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/zero"
)

// NewPoller creates a new Poller that controls an Actor.
//
// It produces new values at an interval of n.
func NewPoller(a Actor, n time.Duration, opts ...PollerOption) *Poller {
	cfg := PollerOptions{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	p := &Poller{
		act:    a,
		log:    logutil.WithComponent(cfg.Logger, (*Poller)(nil)),
		ticker: time.NewTicker(n),
		recv:   make(chan result),
		stop:   make(chan zero.Struct),
	}
	go p.run()
	return p
}

// PollerWithLogger configures a Poller to write logs with log.
func PollerWithLogger(log *logrus.Entry) PollerOption {
	return func(opt *PollerOptions) { opt.Logger = log }
}

type (
	// A Poller can control an Actor.
	Poller struct {
		act Actor
		log *logrus.Entry

		ticker *time.Ticker
		stop   chan zero.Struct
		recv   chan result

		destructor sync.Once
	}

	// PollerOptions configures a Poller.
	PollerOptions struct {
		Logger *logrus.Entry
	}

	// A PollerOption modifies a PollerOptions.
	PollerOption func(*PollerOptions)

	result struct {
		Value zero.Interface
		Error error
	}
)

// Stop stops the Poller; any values that have yet to be passed to the Actor
// will be dropped.
func (p *Poller) Stop() {
	p.destructor.Do(func() {
		close(p.stop)
		p.ticker.Stop()
	})
}

func (p *Poller) run() {
	go p.captureResults()
	go p.produceResult()
	for range p.ticker.C {
		go p.produceResult()
	}
}

func (p *Poller) produceResult() {
	p.log.Trace("Requesting value from Producer...")
	v, err := p.act.Prod()
	p.recv <- result{
		Value: v,
		Error: err,
	}
}

func (p *Poller) captureResults() {
	for {
		select {
		case <-p.stop:
			return
		case res := <-p.recv:
			p.log.
				WithField("result", res).
				Trace("Received result from Producer.")
			p.act.Recv(res.Value, res.Error)
		}
	}
}
