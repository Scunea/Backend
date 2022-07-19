package utils

import (
	"os"

	"github.com/Scunea/Backend/types"
)

func FileExists(file string) bool {
	f, err := os.OpenFile("./files/"+file, os.O_RDONLY, 0600)
	if err == nil {
		f.Close()
	}
	return err != nil
}

func FilesExist(files []types.IdAndName) bool {
	var filesExist = true

	for _, file := range files {
		f, err := os.OpenFile("./files/"+file.Id, os.O_RDONLY, 0600)
		if err == nil {
			f.Close()
		} else {
			filesExist = false
			break
		}
	}

	return filesExist
}
