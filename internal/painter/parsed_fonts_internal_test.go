package painter

import (
	"runtime"
	"testing"
	"weak"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// testFont is a font resource no other cache in the package has seen.
func testFont(name string) fyne.Resource {
	return fyne.NewStaticResource(name, theme.DefaultTextFont().Content())
}

// TestClearFontCache_ReusesFontsStillInUse checks that a font loaded again
// after each settings change is parsed once, not once per change.
func TestClearFontCache_ReusesFontsStillInUse(t *testing.T) {
	res := testFont("reused.ttf")
	ClearFontCache()
	first := loadMeasureFont(res)
	require.NotNil(t, first)

	for range 3 {
		ClearFontCache()
		again := loadMeasureFont(res)
		require.NotNil(t, again)
		assert.Same(t, first.Font, again.Font, "the font was parsed again after a clear")
	}
}

// TestClearFontCache_ReleasesFontsNoLongerUsed checks that a font nothing has
// loaded since the last clear is let go at the next one, so fonts a theme
// stopped using do not stay parsed for good.
func TestClearFontCache_ReleasesFontsNoLongerUsed(t *testing.T) {
	res := testFont("dropped.ttf")
	ClearFontCache()
	face := loadMeasureFont(res)
	require.NotNil(t, face)
	parsed := weak.Make(face.Font)
	face = nil

	ClearFontCache()
	ClearFontCache()
	for i := 0; i < 5 && parsed.Value() != nil; i++ {
		runtime.GC()
	}
	assert.Nil(t, parsed.Value(), "a font unused for a whole clear is still parsed")
	runtime.KeepAlive(res)
}
