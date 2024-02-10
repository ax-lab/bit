package common

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func ReadText(filename string) string {
	out, err := os.ReadFile(filename)
	if err != nil && !os.IsNotExist(err) {
		NoError(err, "reading file text")
	}
	return string(out)
}

func ReadJson(filename string, output any) any {
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		NoError(err, "reading JSON file")
	}

	if output == nil {
		output = &output
	}

	err = json.Unmarshal(data, output)
	NoError(err, "decoding JSON file")
	return output
}

func WriteText(filepath string, text string) {
	if !strings.HasSuffix(text, "\n") {
		text += "\n"
	}
	err := os.WriteFile(filepath, ([]byte)(text), fs.ModePerm)
	NoError(err, "WriteTextIf failed")

}

func WriteJson(filepath string, data any) {
	json, err := json.MarshalIndent(data, "", "    ")
	NoError(err, "WriteJson serialization failed")
	WriteText(filepath, string(json))
}

func PathRelative(base, path string) string {
	fullBase, err := filepath.Abs(base)
	NoError(err, "getting absolute base path for relative")

	fullPath, err := filepath.Abs(path)
	NoError(err, "getting absolute path for relative")

	rel, err := filepath.Rel(fullBase, fullPath)
	NoError(err, "getting relative path")
	return rel
}

func PathWithExtension(filename string, ext string) string {
	out := strings.TrimSuffix(filename, filepath.Ext(filename))
	return out + ext
}
