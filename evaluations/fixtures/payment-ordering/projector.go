package paymentordering

import "errors"

type Event string

const (
	Captured Event = "captured"
	Refunded Event = "refunded"
)

type Projector struct {
	captured bool
	refunded bool
	pending  []Event
}

func (p *Projector) Apply(event Event) error {
	if event == Refunded && !p.captured {
		return errors.New("invalid transition")
	}
	if event == Captured {
		p.captured = true
	}
	if event == Refunded {
		p.refunded = true
	}
	return nil
}

func (p *Projector) State() (captured, refunded bool, pending int) {
	return p.captured, p.refunded, len(p.pending)
}
