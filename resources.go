package kyugo

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
)

var loaded = make(map[string]map[string]string)

// flattenMap converts nested maps into dot-separated keys.
func flattenMap(prefix string, in map[string]interface{}, out map[string]string) {
	for k, v := range in {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		switch vv := v.(type) {
		case string:
			out[key] = vv
		case map[string]interface{}:
			flattenMap(key, vv, out)
		default:
			// try to marshal non-string leaves
			if b, err := json.Marshal(vv); err == nil {
				out[key] = string(b)
			}
		}
	}
}

// LoadFromFS loads language files from the provided filesystem. The fs should
// be rooted at the languages directory (eg. os.DirFS("resources/langs")).
// It reads each subdirectory (language) and merges all JSON files into a
// flattened map for that language.
func LoadFromFS(fsys fs.FS) error {
	if fsys == nil {
		return fmt.Errorf("nil fs provided")
	}
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		lang := e.Name()
		langMap := make(map[string]string)
		// read files inside language dir
		files, err := fs.ReadDir(fsys, lang)
		if err != nil {
			continue
		}
		for _, fi := range files {
			if fi.IsDir() {
				continue
			}
			p := path.Join(lang, fi.Name())
			b, err := fs.ReadFile(fsys, p)
			if err != nil {
				continue
			}
			var obj map[string]interface{}
			if err := json.Unmarshal(b, &obj); err != nil {
				continue
			}
			// flatten under file base name (without extension) to preserve structure
			base := fi.Name()
			// strip extension
			if ext := path.Ext(base); ext != "" {
				base = base[:len(base)-len(ext)]
			}
			flattenMap(base, obj, langMap)
		}
		loaded[lang] = langMap
	}
	return nil
}

// GetAll returns the flattened messages map for the given language, or nil.
func GetAll(lang string) map[string]string {
	return loaded[lang]
}
