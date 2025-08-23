/*
This file is part of Scunea.
Copyright (C) 2025  Mateo Hideg

Scunea is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

Scunea is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with Scunea.  If not, see <https://www.gnu.org/licenses/>.
*/

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
