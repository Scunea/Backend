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

type User struct {
	Token          string `json:"token"`
	Id             string `json:"id"`
	Email          string `json:"email"`
	Verified       bool   `json:"verified"`
	Verificator    string `json:"verificator"`
	Tfa            string `json:"tfa"`
	Name           string `json:"name"`
	Grades         string `json:"grades"`
	Password       string `json:"password"`
	Administrator  string `json:"administrator"`
	Teacher        string `json:"teacher"`
	Parents        string `json:"parents"`
	PendingParents string `json:"pendingparents"`
	Schools        string `json:"schools"`
	PendingSchools string `json:"pendingschools"`
}

type School struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type Activity struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Files       string `json:"files"`
	Type        string `json:"type"`
	Delivery    string `json:"delivery"`
	Author      string `json:"author"`
	Date        string `json:"date"`
	Expiration  string `json:"expiration"`
	Receiver    string `json:"receiver"`
	Delivered   string `json:"delivered"`
	Result      string `json:"result"`
	Viewed      string `json:"viewed"`
	School      string `json:"school"`
}

type Message struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Files    string `json:"files"`
	Author   string `json:"author"`
	Date     string `json:"date"`
	Receiver string `json:"receiver"`
	School   string `json:"school"`
}

type Notification struct {
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
	Id       string `json:"id"`
}

type Report struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	File   string `json:"file"`
	Author string `json:"author"`
	Date   string `json:"date"`
	School string `json:"school"`
}
