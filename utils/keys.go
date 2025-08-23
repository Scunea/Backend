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
	"crypto/rand"
	"log"
	"os"
)

func CreateHmacKey() []byte {
	file, err := os.ReadFile("private.key")
	if err != nil {
		var key = make([]byte, 64)
		rand.Reader.Read(key)
		err = os.WriteFile("private.key", key, 0600)
		if err != nil {
			log.Fatal(err)
		}
		return key
	} else {
		return file
	}
}
