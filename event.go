package gt3

import (
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// Event handling

type EventHandler interface {
	Event(event Event, when time.Time)
}

type EventHandlerFn func(Event, time.Time)

func (fn EventHandlerFn) Event(e Event, t time.Time) { fn(e, t) }

func SetEventCallbacks(w *glfw.Window, handler EventHandler, eventTypes ...Event) {
	s := &eventProvider{handler}
	for _, e := range eventTypes {
		switch e.(type) {
		case RefreshEvent:
			w.SetRefreshCallback(s.postRefreshEvent)
		case CharModsEvent:
			w.SetCharModsCallback(s.postCharModsEvent)
		case CursorEnterEvent:
			w.SetCursorEnterCallback(s.postCursorEnterEvent)
		case CursorPosEvent:
			w.SetCursorPosCallback(s.postCursorPosEvent)
		case DropEvent:
			w.SetDropCallback(s.postDropEvent)
		case FramebufferSizeEvent:
			w.SetFramebufferSizeCallback(s.postFramebufferSizeEvent)
		case IconifyEvent:
			w.SetIconifyCallback(s.postIconifyEvent)
		case KeyEvent:
			w.SetKeyCallback(s.postKeyEvent)
		case MouseEvent:
			w.SetMouseButtonCallback(s.postMouseEvent)
		case CharEvent:
			w.SetCharCallback(s.postCharEvent)
		case CloseEvent:
			w.SetCloseCallback(s.postCloseEvent)
		case FocusEvent:
			w.SetFocusCallback(s.postFocusEvent)
		case PositionEvent:
			w.SetPosCallback(s.postPositionEvent)
		case ResizeEvent:
			w.SetSizeCallback(s.postResizeEvent)
		case ScrollEvent:
			w.SetScrollCallback(s.postScrollEvent)
		}
	}
}

func ClearEventCallbacks(w *glfw.Window) {
	w.SetRefreshCallback(nil)
	w.SetCharModsCallback(nil)
	w.SetCursorEnterCallback(nil)
	w.SetCursorPosCallback(nil)
	w.SetDropCallback(nil)
	w.SetFramebufferSizeCallback(nil)
	w.SetIconifyCallback(nil)
	w.SetKeyCallback(nil)
	w.SetMouseButtonCallback(nil)
	w.SetCharCallback(nil)
	w.SetCloseCallback(nil)
	w.SetFocusCallback(nil)
	w.SetPosCallback(nil)
	w.SetSizeCallback(nil)
	w.SetScrollCallback(nil)
}

// Event types
type (
	Event interface {
		isEvent()
	}

	RefreshEvent struct {
		Window *glfw.Window
	}

	CharModsEvent struct {
		Window *glfw.Window
		Char   rune
		Mods   glfw.ModifierKey
	}

	CursorEnterEvent struct {
		Window  *glfw.Window
		Entered bool
	}

	CursorPosEvent struct {
		Window *glfw.Window
		X      float64
		Y      float64
	}

	DropEvent struct {
		Window *glfw.Window
		Names  []string
	}

	FramebufferSizeEvent struct {
		Window *glfw.Window
		Width  int
		Height int
	}

	IconifyEvent struct {
		Window    *glfw.Window
		Iconified bool
	}

	KeyEvent struct {
		Window *glfw.Window
		Key    glfw.Key
		Code   int
		Action glfw.Action
		Mods   glfw.ModifierKey
	}

	MouseEvent struct {
		Window *glfw.Window
		Button glfw.MouseButton
		Action glfw.Action
		Mods   glfw.ModifierKey
	}

	CharEvent struct {
		Window *glfw.Window
		Char   rune
	}

	CloseEvent struct {
		Window *glfw.Window
	}

	FocusEvent struct {
		Window  *glfw.Window
		Focused bool
	}

	PositionEvent struct {
		Window *glfw.Window
		X      int
		Y      int
	}

	ResizeEvent struct {
		Window *glfw.Window
		Width  int
		Height int
	}

	ScrollEvent struct {
		Window *glfw.Window
		XOff   float64
		YOff   float64
	}
)

func (RefreshEvent) isEvent()         {}
func (CharModsEvent) isEvent()        {}
func (CursorEnterEvent) isEvent()     {}
func (CursorPosEvent) isEvent()       {}
func (DropEvent) isEvent()            {}
func (FramebufferSizeEvent) isEvent() {}
func (IconifyEvent) isEvent()         {}
func (KeyEvent) isEvent()             {}
func (MouseEvent) isEvent()           {}
func (CharEvent) isEvent()            {}
func (CloseEvent) isEvent()           {}
func (FocusEvent) isEvent()           {}
func (PositionEvent) isEvent()        {}
func (ResizeEvent) isEvent()          {}
func (ScrollEvent) isEvent()          {}

// Event provider (hook)

type eventProvider struct {
	events EventHandler
}

func (p *eventProvider) event(e Event) {
	if e != nil {
		p.events.Event(e, time.Now())
	}
}

func (p *eventProvider) postRefreshEvent(Window *glfw.Window) {
	p.event(RefreshEvent{Window})
}

func (p *eventProvider) postCharModsEvent(Window *glfw.Window, Char rune, Mods glfw.ModifierKey) {
	p.event(CharModsEvent{Window, Char, Mods})
}

func (p *eventProvider) postCursorEnterEvent(Window *glfw.Window, Entered bool) {
	p.event(CursorEnterEvent{Window, Entered})
}

func (p *eventProvider) postCursorPosEvent(Window *glfw.Window, X float64, Y float64) {
	p.event(CursorPosEvent{Window, X, Y})
}

func (p *eventProvider) postDropEvent(Window *glfw.Window, Names []string) {
	p.event(DropEvent{Window, Names})
}

func (p *eventProvider) postFramebufferSizeEvent(Window *glfw.Window, Width int, Height int) {
	p.event(FramebufferSizeEvent{Window, Width, Height})
}

func (p *eventProvider) postIconifyEvent(Window *glfw.Window, Iconified bool) {
	p.event(IconifyEvent{Window, Iconified})
}

func (p *eventProvider) postKeyEvent(Window *glfw.Window, Key glfw.Key, Code int, Action glfw.Action, Mods glfw.ModifierKey) {
	p.event(KeyEvent{Window, Key, Code, Action, Mods})
}

func (p *eventProvider) postMouseEvent(Window *glfw.Window, Button glfw.MouseButton, Action glfw.Action, Mods glfw.ModifierKey) {
	p.event(MouseEvent{Window, Button, Action, Mods})
}

func (p *eventProvider) postCharEvent(Window *glfw.Window, Char rune) {
	p.event(CharEvent{Window, Char})
}

func (p *eventProvider) postCloseEvent(Window *glfw.Window) {
	p.event(CloseEvent{Window})
}

func (p *eventProvider) postFocusEvent(Window *glfw.Window, Focused bool) {
	p.event(FocusEvent{Window, Focused})
}

func (p *eventProvider) postPositionEvent(Window *glfw.Window, X int, Y int) {
	p.event(PositionEvent{Window, X, Y})
}

func (p *eventProvider) postResizeEvent(Window *glfw.Window, Width int, Height int) {
	p.event(ResizeEvent{Window, Width, Height})
}

func (p *eventProvider) postScrollEvent(Window *glfw.Window, XOff float64, YOff float64) {
	p.event(ScrollEvent{Window, XOff, YOff})
}
