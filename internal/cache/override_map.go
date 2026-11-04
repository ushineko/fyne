package cache

import (
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
	"weak"

	"fyne.io/fyne/v2"
)

// overrideMap holds the theme scope of each object a theme override covers.
//
// It holds the objects weakly. Renderer expiry removes a widget's entry, but
// nothing removes a canvas primitive's: a Text, Rectangle or Image inside an
// override would otherwise be kept, with its scope, for the life of the
// program. An entry goes when its object is collected instead.
//
// It is safe for concurrent use whatever the build tags, because cleanups run
// on a goroutine of their own.
type overrideMap struct {
	weak sync.Map // objectID → *overrideEntry

	// strong holds objects that are not pointers, which cannot be held weakly.
	strong sync.Map // fyne.CanvasObject → *overrideScope
}

// objectID names an object by its type and address. The type separates a
// struct from a field at its start, which shares its address.
type objectID struct {
	typ  reflect.Type
	addr uintptr
}

type overrideEntry struct {
	obj   weak.Pointer[byte]
	scope atomic.Pointer[overrideScope] // nil once deleted
}

// idOf names o, and returns false for an object that cannot be held weakly.
func idOf(o fyne.CanvasObject) (objectID, unsafe.Pointer, bool) {
	v := reflect.ValueOf(o)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Type().Elem().Size() == 0 {
		return objectID{}, nil, false
	}
	p := v.UnsafePointer()
	return objectID{typ: v.Type(), addr: uintptr(p)}, p, true
}

// entry is o's entry, or nil if it has none. An entry whose object is gone
// belongs to an earlier object at the same address, before its cleanup ran.
func (m *overrideMap) entry(o fyne.CanvasObject) (*overrideEntry, bool) {
	id, _, ok := idOf(o)
	if !ok {
		return nil, false
	}
	e, found := m.weak.Load(id)
	if !found {
		return nil, true
	}
	en := e.(*overrideEntry)
	if en.obj.Value() == nil {
		return nil, true
	}
	return en, true
}

// Load returns the scope o is drawn in.
func (m *overrideMap) Load(o fyne.CanvasObject) (*overrideScope, bool) {
	en, weakly := m.entry(o)
	if !weakly {
		s, ok := m.strong.Load(o)
		if !ok {
			return nil, false
		}
		return s.(*overrideScope), true
	}
	if en == nil {
		return nil, false
	}
	s := en.scope.Load()
	return s, s != nil
}

// Store sets the scope o is drawn in.
func (m *overrideMap) Store(o fyne.CanvasObject, s *overrideScope) {
	id, p, ok := idOf(o)
	if !ok {
		m.strong.Store(o, s)
		return
	}
	if en, _ := m.entry(o); en != nil {
		en.scope.Store(s)
		return
	}

	en := &overrideEntry{obj: weak.Make((*byte)(p))}
	en.scope.Store(s)
	m.weak.Store(id, en)
	runtime.AddCleanup((*byte)(p), func(c overrideCleanup) {
		m.weak.CompareAndDelete(c.id, c.entry)
	}, overrideCleanup{id: id, entry: en})
}

// overrideCleanup is what a collected object's cleanup needs. It holds the
// object only weakly, or the object would never be collected.
type overrideCleanup struct {
	id    objectID
	entry *overrideEntry
}

// Delete removes o's scope. The entry stays until o is collected, so that
// storing a scope again adds no second cleanup.
func (m *overrideMap) Delete(o fyne.CanvasObject) {
	en, weakly := m.entry(o)
	if !weakly {
		m.strong.Delete(o)
		return
	}
	if en != nil {
		en.scope.Store(nil)
	}
}

// Len counts the entries held, whether or not their scope was deleted.
func (m *overrideMap) Len() (count int) {
	for _, s := range []*sync.Map{&m.weak, &m.strong} {
		s.Range(func(_, _ any) bool {
			count++
			return true
		})
	}
	return count
}
