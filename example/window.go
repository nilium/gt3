package main

import (
	"fmt"
	"log"
	"math"
	"runtime"
	"time"

	"go.spiff.io/gt3"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

func init() {
	runtime.LockOSThread()
}

func logstack(msg ...interface{}) {
	var buf [8192]byte
	n := runtime.Stack(buf[:], false)
	b := buf[:n]
	log.Printf("%s\n-!- TRACE -!-\n%s\n-!- END TRACE -!-", fmt.Sprint(msg...), b)
}

type queuedEvent struct {
	e gt3.Event
	t time.Time
}

type eventQueue []queuedEvent

func (q *eventQueue) Event(e gt3.Event, t time.Time) {
	*q = append(*q, queuedEvent{e, t})
}

func (q *eventQueue) play(h gt3.EventHandler) {
	base := *q
	for i, e := range base {
		base[i] = queuedEvent{}
		h.Event(e.e, e.t)
	}
	*q = base[:0]
}

func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}

	if err := gl.Init(); err != nil {
		panic(err)
	}

	down := make(chan struct{})
	sim := gt3.NewSim(2, 30, down)

	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, 1)
	wnd, err := glfw.CreateWindow(800, 600, "Test", nil, nil)
	if err != nil {
		panic(err)
	}

	focused := true
	var queue eventQueue
	var handler gt3.EventHandlerFn = func(ev gt3.Event, _ time.Time) {
		switch ev := ev.(type) {
		case gt3.FocusEvent:
			focused = ev.Focused
			fps := 30
			if !focused {
				fps = 5
			}
			sim.SetRenderFPS(fps)
		case gt3.CloseEvent:
			if down != nil {
				close(down)
				down = nil
				log.Println("Window closed")
			}
		case gt3.KeyEvent:
			if ev.Key == glfw.KeyEscape && ev.Action == glfw.Release && down != nil {
				close(down)
				down = nil
				log.Println("Escape pressed")
			}
		default:
			log.Printf("Unrecognized event %#+v", ev)
		}
	}
	gt3.SetEventCallbacks(wnd, &queue, gt3.FocusEvent{}, gt3.CloseEvent{}, gt3.KeyEvent{})

	sim.PreFrame = gt3.OpFn(func(step, ft float64, rt time.Time) {
		glfw.PollEvents()

		queue.play(handler)

		if !focused {
			time.Sleep(time.Millisecond * 11)
		}
	})

	sim.Frame = gt3.OpFn(func(step, ft float64, rt time.Time) {
		log.Print("Frame  | ft=", ft, "->", ft+step, " rt=", rt)
	})

	sim.Render = gt3.OpFn(func(step, ft float64, rt time.Time) {
		cur := sim.Seconds() - step
		log.Print("Render | ft=", ft, " st=", cur, "->", cur+step, " rt=", rt)

		wnd.MakeContextCurrent()

		// Just make it clear the context is working.
		r := float32(math.Abs(math.Sin(sim.Now() * 0.2)))
		gl.ClearColor(r, 0.3, 0.3, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		// No pun intended.

		wnd.SwapBuffers()
	})

	sim.Run()
}
