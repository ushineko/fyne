package painter

import (
	"image"
	"image/color"
	"runtime"
	"testing"
	"weak"

	"github.com/go-text/typesetting/font"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type nopAtlasTexture struct{}

func (nopAtlasTexture) FlushGlyphs()                      {}
func (nopAtlasTexture) UploadGlyph(*image.RGBA, int, int) {}
func (nopAtlasTexture) ResizeAtlas(int)                   {}

// TestGlyphAtlas_ReleasesFacesAfterFontCacheClear checks that an atlas which
// drew with a face lets it go once the font cache is cleared and it draws
// again. Each face holds its parsed font, so a kept one is a font set kept
// per theme or settings change.
func TestGlyphAtlas_ReleasesFacesAfterFontCacheClear(t *testing.T) {
	ClearFontCache()
	var a GlyphAtlas
	text := canvas.NewText("Atlas text", color.White)

	old := atlasFace(t, &a, text)
	ClearFontCache()
	atlasFace(t, &a, text)

	for i := 0; i < 5 && old.Value() != nil; i++ {
		runtime.GC()
	}
	assert.Nil(t, old.Value(), "a face from before the clear is still reachable")
	runtime.KeepAlive(&a) // the atlas outlives the check, as a painter's does
}

// atlasFace draws text through a and returns a weak pointer to the face its
// glyphs came from.
func atlasFace(t *testing.T, a *GlyphAtlas, text *canvas.Text) weak.Pointer[font.Face] {
	t.Helper()
	_, ok := a.TextQuads(nil, text, fyne.NewPos(0, 0), 1, -1e9, 1e9, nopAtlasTexture{})
	require.True(t, ok)
	glyphs, ok := a.shape(text, 1)
	require.True(t, ok)
	require.NotEmpty(t, glyphs)
	return weak.Make(glyphs[0].Face)
}
