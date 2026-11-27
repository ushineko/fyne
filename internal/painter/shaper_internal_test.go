package painter

import (
	"runtime"
	"testing"
	"weak"

	"github.com/go-text/typesetting/font"
	"github.com/stretchr/testify/assert"

	"fyne.io/fyne/v2"
)

// TestClearFontCache_ReleasesShapedFaces checks that a face which has shaped
// text can be collected once the font cache is cleared. Every settings change
// clears the cache and loads new faces, so anything that keeps the old ones
// grows by a parsed font set per change.
func TestClearFontCache_ReleasesShapedFaces(t *testing.T) {
	ClearFontCache()
	old := shapedFace(t)

	ClearFontCache()
	shapedFace(t) // the faces that replace it shape text too
	for i := 0; i < 5 && old.Value() != nil; i++ {
		runtime.GC()
	}
	assert.Nil(t, old.Value(), "a face from before the clear is still reachable")
}

// shapedFace shapes a string with the app's regular face and returns a weak
// pointer to that face, holding no strong reference of its own.
func shapedFace(t *testing.T) weak.Pointer[font.Face] {
	t.Helper()
	item := CachedFontFace(fyne.TextStyle{}, nil, nil)
	MeasureString(item.Fonts, "Shaped text", 14, fyne.TextStyle{})
	faces := item.Fonts.(*dynamicFontMap).faces
	if len(faces) == 0 {
		t.Fatal("no face loaded")
	}
	return weak.Make(faces[0])
}
