package template

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"reflect"
	"strings"
	"text/template"
)

var (
	ErrValidation  = fmt.Errorf("strict checking: template validation failed")
	ErrMissingData = fmt.Errorf("source data value not present")
)

// Assets represents a static directory which contains assets and templates
type Assets struct {
	dir []fs.DirEntry
	fs  fs.FS
}

// NewAssets provides a new Assets object when given a filesystem and a directory name.
//
// For a webserver, this filesystem will probably point to a 'static' directory.
//
// Filesystem can be either a local filesystem (as expected in development),
// or an embed.FS (as expected for production use cases).
func NewAssets(filesystem fs.FS, directory string) (*Assets, error) {
	dir, err := fs.ReadDir(filesystem, directory)
	if err != nil {
		return nil, fmt.Errorf("open static dir: %w", err)
	}
	return &Assets{dir: dir, fs: filesystem}, nil
}

// Make executes a set of templates (defined by their path) and injects arbitrary data.
// When 'strict' is true, the rendered template is checked for the string representation
// of every value in 'data'. If a value is not found, an ErrValidation is returned.
//
// Template paths must be provided in the order they will be evaluated; parent -> child.
func (h *Assets) Make(templatePaths []string, data interface{}, strict bool) (string, error) {
	// todo can we not keep the filesystem obj since it's already parsed in New?
	tmpl, err := template.ParseFS(h.fs, templatePaths...)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}
	// If a child template is called before a parent template, the output will be empty
	len := buf.Len()
	if len == 0 {
		return "", fmt.Errorf("template output is empty, ensure parents are called before children")
	}

	validated := 0
	if strict {
		content, err := parseContent(data)
		if err != nil {
			return "", fmt.Errorf("strict check, parse content: %w", err)
		}
		for k, v := range content {
			exptectedValue, ok := v.(string)
			if !ok {
				return "", fmt.Errorf("%w: k/v pair not parsable as string", ErrValidation)
			}
			if !strings.Contains(buf.String(), exptectedValue) {
				return "", fmt.Errorf("%w: %w: key %q", ErrValidation, ErrMissingData, k)
			}
			validated++
			// slog.Debug("strict check: data value succesfully found in rendered template", "key", k)
		}
	}
	slog.Info("template(s) executed",
		"bytes", len,
		"templates", templatePaths,
		"strict", strict,
		"fields_validated", validated,
	)
	return buf.String(), nil
}

// parseContent parses a nested data structure into a flat map for validationg template output.
func parseContent(data interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	err := parseRecursive(data, "", result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// parseRecursive parses a nested data structure into a flat map, and uses cursed reflection.
func parseRecursive(data interface{}, prefix string, result map[string]interface{}) error {
	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Ptr:
		return parseRecursive(val.Elem().Interface(), prefix, result)
	case reflect.Struct:
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			if !field.IsExported() {
				// skip quietely here; rely on 'strict' checking to catch the omission,
				// because an unexported field that is also missing from the template
				// is none of our business.
				continue
			}
			fieldValue := val.Field(i).Interface()
			key := prefix + field.Name
			err := parseRecursive(fieldValue, key+".", result)
			if err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range val.MapKeys() {
			mapValue := val.MapIndex(key).Interface()
			mapKey := fmt.Sprintf("%s%v", prefix, key)
			err := parseRecursive(mapValue, mapKey+".", result)
			if err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i).Interface()
			key := fmt.Sprintf("%s[%d]", prefix, i)
			err := parseRecursive(elem, key+".", result)
			if err != nil {
				return err
			}
		}
	default:
		// For other types, add the value to the map
		result[prefix] = fmt.Sprintf("%v", val.Interface())
	}
	return nil
}
