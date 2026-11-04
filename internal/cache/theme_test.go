package cache

import (
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
)

// plainObject stands in for a canvas primitive: an object that is not a widget,
// so no renderer expiry ever removes its override.
type plainObject struct {
	fyne.CanvasObject
	n int
}

type dummyTheme struct {
	fyne.Theme
	n int
}

// collectUntil runs the collector until done reports true or a second passes,
// as cleanups run on their own goroutine some time after a collection.
func collectUntil(done func() bool) {
	deadline := time.Now().Add(time.Second)
	for !done() && time.Now().Before(deadline) {
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}
}

func TestAnOverrideKeepsNoObjectThatIsGone(t *testing.T) {
	before := overrides.Len()
	for range 100 {
		OverrideTheme(&fyne.Container{Objects: []fyne.CanvasObject{&plainObject{}}}, &dummyTheme{})
	}

	collectUntil(func() bool { return overrides.Len()-before < 10 })
	assert.Less(t, overrides.Len()-before, 10)
}

func TestAnOverriddenObjectKeepsItsThemeWhileItLives(t *testing.T) {
	o := &plainObject{}
	th := &dummyTheme{}
	OverrideTheme(&fyne.Container{Objects: []fyne.CanvasObject{o}}, th)

	runtime.GC()
	runtime.GC()
	assert.Same(t, th, WidgetTheme(o))
	runtime.KeepAlive(o)
}

func TestAnObjectOverriddenAgainTakesTheNewTheme(t *testing.T) {
	o := &plainObject{}
	OverrideTheme(&fyne.Container{Objects: []fyne.CanvasObject{o}}, &dummyTheme{})
	th := &dummyTheme{n: 1}
	OverrideTheme(&fyne.Container{Objects: []fyne.CanvasObject{o}}, th)

	assert.Same(t, th, WidgetTheme(o))
}

func TestADeletedOverrideIsGoneAndCanBeSetAgain(t *testing.T) {
	o := &dummyWidget{}
	th := &dummyTheme{}
	OverrideTheme(o, th)

	overrides.Delete(o)
	assert.Nil(t, WidgetTheme(o))
	assert.Empty(t, WidgetScopeID(o))

	OverrideTheme(o, th)
	assert.Same(t, th, WidgetTheme(o))
}
