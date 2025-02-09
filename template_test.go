package template

import (
	"embed"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var staticTestDir = path.Join("static")

// an embed filesystem to test expected production use cases
//
//go:embed static
var static embed.FS

type planet struct {
	Name       string
	Distance   float64 // million km
	HasRings   bool
	Moons      int
	Atmosphere []string
	Attributes map[string]string
	Discovery  time.Time
}

type footer struct {
	Year int
}

type page struct {
	Title  string
	Body   planet
	Footer footer
}

var jupiter = planet{
	Name:       "Jupiter",
	Distance:   778.5,
	HasRings:   true,
	Moons:      79,
	Atmosphere: []string{"Hydrogen", "Helium"},
	Attributes: map[string]string{"Diameter": "142,984 km", "Mass": "1.898 × 10^27 kg"},
	Discovery:  time.Date(1610, time.January, 7, 0, 0, 0, 0, time.UTC),
}

var FooterVar = footer{
	Year: time.Now().Year(),
}

var pageVar = page{
	Title:  "My Favorite Planet",
	Body:   jupiter,
	Footer: FooterVar,
}

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func TestReadFS(t *testing.T) {
	t.Run("local FS", func(t *testing.T) {
		entries, err := fs.ReadDir(os.DirFS(staticTestDir), ".")
		require.NoError(t, err)
		require.NotNil(t, entries)

		for _, file := range entries {
			if file.IsDir() {
				t.Logf("dir: %s", file.Name())
				continue
			}
			t.Logf("file: %s", file.Name())
		}
	})

	t.Run("embed FS", func(t *testing.T) {
		entries, err := fs.ReadDir(static, "static")
		require.NoError(t, err)
		require.NotNil(t, entries)

		for _, file := range entries {
			if file.IsDir() {
				t.Logf("dir: %s", file.Name())
				continue
			}
			t.Logf("file: %s", file.Name())
		}
	})
}

func TestMake(t *testing.T) {
	// notes: order matters! parent -> child
	templates := []string{
		"static/html/index.html",
		"static/html/footer.html",
	}

	t.Run("local FS render", func(t *testing.T) {
		assets, err := NewAssets(os.DirFS("."), ".")
		require.NoError(t, err)
		rendered, err := assets.Make(templates, pageVar, true)
		require.NoError(t, err)
		t.Logf("rendered from local filesystem:\n%v\n", rendered)
	})
	t.Run("embed FS render", func(t *testing.T) {
		assets, err := NewAssets(static, "static")
		require.NoError(t, err)
		rendered, err := assets.Make(templates, pageVar, true)
		require.NoError(t, err)
		t.Logf("rendered from embed filesystem\n%v\n", rendered)
	})
}
