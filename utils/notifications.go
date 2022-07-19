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
