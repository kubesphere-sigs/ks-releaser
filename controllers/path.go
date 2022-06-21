package controllers

import (
	"errors"
	"io/fs"
	"path/filepath"
)

func findReleaserFile(name, dir string) (result string) {
	_ = filepath.Walk(dir, func(basepath string, info fs.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == name {
			result = basepath
			return errors.New("found file: " + result)
		}
		return nil
	})
	return
}
