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
	"github.com/Scunea/Backend/webpush"
	"github.com/jmoiron/sqlx"
)

func SendNotification(db *sqlx.DB, user string, message interface{}) {
	var notifications = GetNotifications(db)
	var notificationInfo = FilterFunc(notifications, func(notification types.Notification) bool {
		return notification.Id == user
	})[0]

	if len(notificationInfo.Endpoint) > 0 {
		response, err := webpush.SendNotification([]byte(UnParse(message)), &webpush.Subscription{
			Endpoint: notificationInfo.Endpoint,
			Keys: webpush.Keys{
				Auth:   notificationInfo.Auth,
				P256dh: notificationInfo.P256dh,
			},
		}, &webpush.Options{
			Subscriber:      os.Getenv("EMAIL"),
			VAPIDPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
			VAPIDPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
		})

		if err != nil || response.StatusCode == 404 || response.StatusCode == 410 {
			db.Exec("DELETE FROM notifications WHERE id = $1", user)
		}
	}
}
