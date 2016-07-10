package gt3

import (
	"math"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type Op interface {
	Do(frameTime float64, when time.Time)
}

type OpFn func(frameTime float64, when time.Time)

func (fn OpFn) Do(frameTime float64, when time.Time) { fn(frameTime, when) }

type Sim struct {
	PreFrame Op
	Frame    Op
	Render   Op

	fps  int // Simulation limitation
	rfps int // Rendition limitation

	// Timing
	hz         float64
	rhz        float64
	baseTime   float64
	simTime    float64
	renderTime float64
	runTime    int64

	sched   chan Op
	stopped <-chan struct{}
}

func NewSim(fps, rfps int, stop <-chan struct{}) *Sim {
	if fps <= 0 {
		panic("gt3: simloop FPS must be > 0")
	}

	var rhz float64
	if rfps > 0 {
		rhz = 1.0 / float64(rfps)
	}

	return &Sim{
		fps:  fps,
		rfps: rfps,
		hz:   1.0 / float64(fps),
		rhz:  rhz,

		sched:   make(chan Op),
		stopped: stop,
	}
}

func (s *Sim) now() float64 {
	return glfw.GetTime() - s.baseTime
}

func realtime(unixBase int64, base, after float64) time.Time {
	var (
		real  = base + after
		secs  = unixBase + int64(real)
		frac  = real - math.Floor(real)
		nanos = int64(float64(time.Second) * frac)
	)

	return time.Unix(secs, nanos)
}

func (s *Sim) RealTime() time.Time {
	return realtime(s.runTime, s.baseTime, s.simTime)
}

func (s *Sim) pollSched(ft float64, rt time.Time) {
	for sched := s.sched; ; {
		select {
		case op := <-sched:
			op.Do(ft, rt)
		default:
			return
		}
	}
}

func runOp(op Op, ft float64, rt time.Time) {
	if op != nil {
		op.Do(ft, rt)
	}
}

func (s *Sim) frame(ft float64, rt time.Time) {
	s.pollSched(ft, rt)
	runOp(s.Frame, ft, rt)
}

func (s *Sim) Run() error {
	stopped := s.stopped

	ubase := time.Now().Unix()
	glfw.SetTime(0)

	s.runTime = ubase
	s.simTime, s.baseTime = 0, glfw.GetTime()
	for {
		select {
		case <-stopped:
			return nil
		default:
		}

		var (
			now  = s.now()
			sim  = s.simTime
			base = s.baseTime
		)

		runOp(s.PreFrame, sim, realtime(ubase, base, sim))

		for hz := s.hz; sim+hz <= now; {
			s.frame(sim, realtime(ubase, base, sim))
			sim += hz
			s.simTime = sim
		}

		if s.rfps > 0 {
			// Reacquire current time and see if we're OK to render since the last render time
			now = s.now()
			if rt := s.renderTime; now >= rt {
				runOp(s.Render, sim, realtime(ubase, base, sim))
				s.renderTime = now + s.rhz
			}
		} else {
			runOp(s.Render, sim, realtime(ubase, base, sim))
		}
	}

	return nil
}

// Sched schedules an op to run on the main goroutine. Sched does not wait for the op to run.
func (s *Sim) Sched(op Op) {
	go func() {
		select {
		case s.sched <- op:
		case <-s.stopped:
		}
	}()
}

// Sync schedules an Op to run on the main goroutine and waits for it to finish running. If scheduled on the main
// goroutine, Sync will deadlock the process.
func (s *Sim) Sync(op Op) {
	done := make(chan struct{})
	syncOp := OpFn(func(ft float64, w time.Time) {
		defer close(done)
		op.Do(ft, w)
	})

	select {
	case s.sched <- syncOp:
		<-done
	case <-s.stopped:
	}
}
