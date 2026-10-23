package link

import "time"

const (
	rttMinimumHold         = 10 * time.Second
	rttConfirmationWindow  = 3 * time.Second
	rttMinimumImprovement  = 2.0 // milliseconds
	rttImprovementFraction = 0.20
)

type rttSelection struct {
	candidate string
	since     time.Time
}

// stableRTTLocked dampens automatic RTT optimization only. Initial selection,
// failed/excluded active sessions and explicit preferences bypass this delay.
func (e *Edge) stableRTTLocked(now time.Time, best *Link, currentEligible bool) *Link {
	current := e.links[e.active]
	if !currentEligible || current == nil || best == nil || best == current {
		e.rttSelection = rttSelection{}
		return best
	}
	currentRTT := current.Stats().RTTMillis
	improvement := currentRTT - best.Stats().RTTMillis
	if now.Sub(e.switched) < rttMinimumHold || improvement < max(rttMinimumImprovement, currentRTT*rttImprovementFraction) {
		e.rttSelection = rttSelection{}
		return current
	}
	if e.rttSelection.candidate != best.ID() {
		e.rttSelection = rttSelection{candidate: best.ID(), since: now}
		return current
	}
	if now.Sub(e.rttSelection.since) < rttConfirmationWindow {
		return current
	}
	e.rttSelection = rttSelection{}
	return best
}
