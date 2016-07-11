package gt3

import (
	"errors"
	"math"
	"sync"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type Op interface {
	Do(step, frameTime float64, when time.Time)
}

type OpFn func(step, frameTime float64, when time.Time)

func (fn OpFn) Do(step, frameTime float64, when time.Time) { fn(step, frameTime, when) }

type Sim struct {
	PreFrame Op
	Frame    Op
	Render   Op

	fps  int // Simulation limitation
	hz   float64
	rfps int // Rendition limitation
	rhz  float64

	// Controls access to FPS/hertz variables
	fpsrw sync.RWMutex

	// Timing
	runTime    int64
	baseTime   float64
	simTime    float64
	renderTime float64

	sched   chan Op
	stopped <-chan struct{}
}

func NewSim(fps, renderfps int, stop <-chan struct{}) *Sim {
	if fps <= 0 {
		panic("gt3: simloop FPS must be > 0")
	}

	var rhz float64
	if renderfps > 0 {
		rhz = 1.0 / float64(renderfps)
	}

	return &Sim{
		fps:  fps,
		rfps: renderfps,
		hz:   1.0 / float64(fps),
		rhz:  rhz,

		stopped: stop,
	}
}

var ErrBadFPS = errors.New("gt3: FPS must be > 0")

func (s *Sim) SetRenderFPS(fps int) (previous int) {
	s.fpsrw.Lock()
	defer s.fpsrw.Unlock()

	previous = s.rfps

	s.rfps = fps
	s.rhz = 1.0 / float64(fps)

	return previous
}

func (s *Sim) SetFPS(fps int) (previous int, err error) {
	if fps < 0 {
		return 0, ErrBadFPS
	}

	s.fpsrw.Lock()
	defer s.fpsrw.Unlock()

	previous = s.fps

	s.fps = fps
	s.hz = 1.0 / float64(fps)

	return previous, nil
}

func (s *Sim) Now() float64 {
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

func (s *Sim) Seconds() float64 {
	return s.simTime
}

func (s *Sim) Time() time.Time {
	return realtime(s.runTime, s.baseTime, s.simTime)
}

func (s *Sim) RealTime() time.Time {
	return realtime(s.runTime, s.baseTime, s.Now())
}

func (s *Sim) pollSched(hz, ft float64, rt time.Time) {
	for sched := s.sched; ; {
		select {
		case op := <-sched:
			op.Do(hz, ft, rt)
		default:
			return
		}
	}
}

func runOp(op Op, hz, ft float64, rt time.Time) {
	if op != nil {
		op.Do(hz, ft, rt)
	}
}

func (s *Sim) frame(hz, ft float64, rt time.Time) {
	s.pollSched(hz, ft, rt)
	runOp(s.Frame, hz, ft, rt)
}

var ErrStopped = errors.New("gt3: stopped")

func (s *Sim) runSim(ubase int64, stopped <-chan struct{}) error {
	select {
	case <-stopped:
		return ErrStopped
	default:
	}

	s.fpsrw.RLock()
	var (
		now  float64
		hz   float64
		sim  = s.simTime
		base = s.baseTime
	)
	s.fpsrw.RUnlock()

	// Refresh hz per-frame
	s.fpsrw.RLock()
	hz = s.hz
	s.fpsrw.RUnlock()

	runOp(s.PreFrame, hz, sim, realtime(ubase, base, sim))

	for now = s.Now(); sim < now; now = s.Now() {
		s.frame(hz, sim, realtime(ubase, base, sim))
		sim += hz
		s.simTime = sim

		if sim < now {
			// Refresh hz per-frame
			s.fpsrw.RLock()
			hz = s.hz
			s.fpsrw.RUnlock()
		}
	}

	// Check if we need to limit rendition FPS
	s.fpsrw.RLock()
	var (
		rlimit = s.rfps > 0
		rhz    = s.rhz
	)
	s.fpsrw.RUnlock()

	if rlimit {
		// Reacquire current time and see if we're OK to render since the last render time
		if rt := s.renderTime; now >= rt {
			runOp(s.Render, hz, now, realtime(ubase, base, now))
			s.renderTime = now + rhz
		}
	} else {
		runOp(s.Render, hz, now, realtime(ubase, base, now))
		s.renderTime = now
	}

	return nil
}

func (s *Sim) Run() error {
	stopped := s.stopped

	ubase := time.Now().Unix()
	glfw.SetTime(0)

	s.sched = make(chan Op)
	s.runTime = ubase
	s.simTime, s.baseTime = 0, glfw.GetTime()
	for {
		if err := s.runSim(ubase, stopped); err != nil {
			return err
		}
	}
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
	syncOp := OpFn(func(hz, ft float64, w time.Time) {
		defer close(done)
		op.Do(hz, ft, w)
	})

	select {
	case s.sched <- syncOp:
		<-done
	case <-s.stopped:
	}
}
