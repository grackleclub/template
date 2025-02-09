# template

Template package for go.

## How to Use

1. Create a New `Assets` Object
    - **local** filesystem (testing)
        ```go
        assets, err := NewAssets(os.DirFS("."), ".")
        if err != nil {
            // handle error    
        }
        ```
    - **embed** filesystem (production)
        ```go
        assets, err := NewAssets(static, "static")
        if err != nil {
            // handle error    
        }
        ```

1. Delare your templates in the order they will be parsed. Parents first, then descendents.
    ```go
    templates := []string{
        "static/html/index.html",
        "static/html/footer.html",
    }
    ```

1. Declare create a set of `type`s and `var`s that mirror the template fields in the prior step's templates. A "kitchen sink" nested complex example is shown:
    ```go
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
    ```
    > [!CAUTION]
    > If a field is not exported (capitalized), it cannot be read by template parsing.


4. Render the template using provided data.
    ```go
    rendered, err := assets.Make(templates, pageVar, true)
    if err != nil {
        return err
    }
    // write 'rendered' string to output
    ```


## `go-run`

It is recommended to use live reload with [`go-run`](https://github.com/grackleclub/go-run), a wrapper that watches for changes and initiates an automatic relaunch and rerun if any files are changed. Run without arguments to target the current directory.