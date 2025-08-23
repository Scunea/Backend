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

package types

type Deliver struct {
	Comments string      `json:"comments"`
	Files    []IdAndName `json:"files"`
	Date     int64       `json:"date"`
}

type ActivityParsed struct {
	Id          string             `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Files       []IdAndName        `json:"files"`
	Type        string             `json:"type"`
	Delivery    string             `json:"delivery"`
	Author      IdAndName          `json:"author"`
	Date        int64              `json:"date"`
	Expiration  int64              `json:"expiration"`
	Receiver    []string           `json:"receiver"`
	Delivered   map[string]Deliver `json:"delivered"`
	Result      map[string]string  `json:"result"`
	Viewed      []string           `json:"viewed"`
	School      string             `json:"school"`
}

type ActivityParsedAlone struct {
	Id          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Files       []IdAndName `json:"files"`
	Type        string      `json:"type"`
	Delivery    string      `json:"delivery"`
	Author      IdAndName   `json:"author"`
	Date        int64       `json:"date"`
	Expiration  int64       `json:"expiration"`
	Receiver    []string    `json:"receiver"`
	Delivered   Deliver     `json:"delivered"`
	Result      string      `json:"result"`
	Viewed      bool        `json:"viewed"`
	School      string      `json:"school"`
}
