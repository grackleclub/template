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

var testDir = path.Join("static")

// an embed filesystem to test expected production use cases
//
//go:embed static
var static embed.FS

func init() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
}

func TestReadFS(t *testing.T) {
	t.Run("local FS", func(t *testing.T) {
		entries, err := fs.ReadDir(os.DirFS(testDir), ".")
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
	s, err := NewAssets(static, "static")
	require.NoError(t, err)

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

	rendered, err := s.Make([]string{"static/html/index.html", "static/html/footer.html"}, pageVar, true)
	require.NoError(t, err)
	t.Logf("\n%v\n", rendered)
}
