package kyugo

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"
)

var loaded = make(map[string]map[string]string)

// Resources holds raw file contents for any file under the resources
// directory. Keys are the relative file paths (unix-style) as provided by
// the FS passed to LoadResources.
var Resources = make(map[string][]byte)

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
// LoadResources loads all files from the provided filesystem root into memory
// and also processes the `langs` subfolder as language resources (JSON files).
// The fs should be rooted at the `resources` directory (eg. os.DirFS("resources")).
func LoadResources(fsys fs.FS) error {
	if fsys == nil {
		return fmt.Errorf("nil fs provided")
	}
	// walk the FS and load every file
	err := fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		// normalize to unix-style path
		np := path.Clean(p)
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return nil
		}
		Resources[np] = b

		// if this file is under langs/<lang> and is a .json file, merge into loaded
		if strings.HasPrefix(np, "langs/") && strings.HasSuffix(np, ".json") {
			parts := strings.SplitN(np, "/", 3) // [langs, lang, file]
			if len(parts) >= 2 {
				lang := parts[1]
				var obj map[string]interface{}
				if err := json.Unmarshal(b, &obj); err == nil {
					base := path.Base(np)
					if ext := path.Ext(base); ext != "" {
						base = base[:len(base)-len(ext)]
					}
					langMap := loaded[lang]
					if langMap == nil {
						langMap = make(map[string]string)
					}
					flattenMap(base, obj, langMap)
					loaded[lang] = langMap
				}
			}
		}
		return nil
	})
	return err
}

// GetAll returns the flattened messages map for the given language, or nil.
func GetAll(lang string) map[string]string {
	return loaded[lang]
}

// GetResource returns the raw bytes for a resource path previously loaded
// via LoadResources. The path should be the relative path inside resources
// (for example "docs/index.html" or "langs/en-US/locale.json").
func GetResource(p string) ([]byte, bool) {
	p = path.Clean(p)
	// try exact
	if b, ok := Resources[p]; ok {
		return b, true
	}
	// try under resources/ prefix
	if b, ok := Resources[path.Join("resources", p)]; ok {
		return b, true
	}
	// try with ./ prefix
	if b, ok := Resources[path.Join(".", p)]; ok {
		return b, true
	}
	return nil, false
}
