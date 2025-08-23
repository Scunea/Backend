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

func ExistsFunc[E any](s []E, f func(E) bool) bool {
	for _, v := range s {
		if f(v) {
			return true
		}
	}
	return false
}

func IndexFunc[E any](s []E, f func(E) bool) int {
	for i, v := range s {
		if f(v) {
			return i
		}
	}
	return -1
}

func FilterFunc[E any](s []E, f func(E) bool) []E {
	var filtered []E
	for _, v := range s {
		if f(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

func Remove[E any](s []E, i int) []E {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
