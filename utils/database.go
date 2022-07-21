package utils

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/Scunea/Backend/types"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func InitDb(db *sqlx.DB) {
	tx := db.MustBegin()
	tx.MustExec("CREATE TABLE IF NOT EXISTS users ( token text NOT NULL, id text NOT NULL, email text NOT NULL, verified boolean NOT NULL, verificator text NOT NULL, tfa text NOT NULL, name text NOT NULL, grades text NOT NULL, password text NOT NULL, administrator text NOT NULL, teacher text NOT NULL, parents text NOT NULL, pendingparents text NOT NULL, schools text NOT NULL, pendingschools text NOT NULL, PRIMARY KEY (id) )")
	tx.MustExec("CREATE TABLE IF NOT EXISTS notifications ( endpoint text NOT NULL, p256dh text NOT NULL, auth text NOT NULL, id text NOT NULL, PRIMARY KEY (endpoint) )")
	tx.MustExec("CREATE TABLE IF NOT EXISTS schools ( id text NOT NULL, name text NOT NULL, logo text NOT NULL, PRIMARY KEY (id) )")
	tx.MustExec("CREATE TABLE IF NOT EXISTS messages ( id text NOT NULL, title text NOT NULL, content text NOT NULL, files text NOT NULL, author text NOT NULL, date text NOT NULL, receiver text NOT NULL, school text NOT NULL, PRIMARY KEY (id) )")
	tx.MustExec("CREATE TABLE IF NOT EXISTS reports ( id text NOT NULL, title text NOT NULL, file text NOT NULL, author text NOT NULL, date text NOT NULL, school text NOT NULL, PRIMARY KEY (id) )")
	tx.MustExec("CREATE TABLE IF NOT EXISTS activities ( id text NOT NULL, title text NOT NULL, description text NOT NULL, files text NOT NULL, type text NOT NULL, delivery TEXT NOT NULL, author text NOT NULL, date text NOT NULL, expiration text NOT NULL, receiver text NOT NULL, delivered text NOT NULL, result text NOT NULL, viewed text NOT NULL, school text NOT NULL, PRIMARY KEY (id) )")
	tx.Commit()
}

func GetUsers(db *sqlx.DB) []types.User {
	var users []types.User
	rows, err := db.Queryx("SELECT * FROM users")
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var tmp types.User
		err := rows.StructScan(&tmp)
		if err != nil {
			log.Fatalln(err)
		}
		users = append(users, tmp)
	}
	return users
}

func GetUsersFromSchool(db *sqlx.DB, schoolId string) []types.User {
	var users []types.User
	rows, err := db.Queryx("SELECT * FROM users")
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var tmp types.User
		err := rows.StructScan(&tmp)
		if err != nil {
			log.Fatalln(err)
		}
		if ExistsFunc(ParseStringArray(tmp.Schools), func(school string) bool {
			return school == schoolId
		}) {
			users = append(users, tmp)
		}
	}
	return users
}

func GetUser(db *sqlx.DB, who string, how string) (types.User, bool) {
	rows, err := db.Queryx("SELECT * FROM users WHERE "+how+" = $1", who)
	if err != nil {
		log.Fatalln(err)
	}
	if rows.Next() {
		var user types.User
		err = rows.StructScan(&user)
		if err != nil {
			log.Fatalln(err)
		}
		rows.Close()
		return user, true
	} else {
		return types.User{}, false
	}
}

func GetMyUser(db *sqlx.DB, c *gin.Context) types.User {
	rows, err := db.Queryx("SELECT * FROM users WHERE token = $1", strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", ""))
	if err != nil {
		log.Fatalln(err)
	}
	if rows.Next() {
		var user types.User
		err = rows.StructScan(&user)
		if err != nil {
			log.Fatalln(err)
		}
		rows.Close()
		return user
	} else {
		return types.User{}
	}
}

func GetChildren(db *sqlx.DB, parent types.User) []types.User {
	var users = GetUsers(db)
	return FilterFunc(users, func(user types.User) bool {
		var parsedParents = ParseStringArray(user.Parents)
		return ExistsFunc(parsedParents, func(parsedParent string) bool {
			return parsedParent == parent.Id
		})
	})
}

func GetSchool(db *sqlx.DB, schoolId string) (types.School, bool) {
	var school types.School
	rows, err := db.Queryx("SELECT * FROM schools WHERE id = $1", schoolId)
	if err != nil {
		log.Fatalln(err)
	}
	if rows.Next() {
		err = rows.StructScan(&school)
		if err != nil {
			log.Fatalln(err)
		}
		rows.Close()
		return school, true
	} else {
		return types.School{}, false
	}
}

func GetSchoolsWithName(db *sqlx.DB, jsonString string) []types.School {
	var arr []string
	json.Unmarshal([]byte(jsonString), &arr)
	var schools []types.School
	for _, school := range arr {
		var schoolObj, ok = GetSchool(db, school)
		if ok {
			schools = append(schools, schoolObj)
		}
	}
	return schools
}

func GetActivities(db *sqlx.DB, schoolId string) []types.Activity {
	var activities []types.Activity
	rows, err := db.Queryx("SELECT * FROM activities WHERE school = $1", schoolId)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var activity types.Activity
		err := rows.StructScan(&activity)
		if err != nil {
			log.Fatalln(err)
		}
		activities = append(activities, activity)
	}
	return activities
}

func GetActivity(db *sqlx.DB, activityId string, schoolId string) (types.Activity, bool) {
	rows, err := db.Queryx("SELECT * FROM activities WHERE school = $1 AND id = $2", schoolId, activityId)
	if err != nil {
		log.Fatalln(err)
	}
	if rows.Next() {
		var activity types.Activity
		err = rows.StructScan(&activity)
		if err != nil {
			log.Fatalln(err)
		}
		rows.Close()
		return activity, true
	} else {
		return types.Activity{}, false
	}
}

func GetMessages(db *sqlx.DB, schoolId string) []types.Message {
	var messages []types.Message
	rows, err := db.Queryx("SELECT * FROM messages WHERE school = $1", schoolId)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var message types.Message
		err := rows.StructScan(&message)
		if err != nil {
			log.Fatalln(err)
		}
		messages = append(messages, message)
	}
	return messages
}

func GetMessage(db *sqlx.DB, messageId string, schoolId string) (types.Message, bool) {
	rows, err := db.Queryx("SELECT * FROM messages WHERE school = $1 AND id = $2", schoolId, messageId)
	if err != nil {
		log.Fatalln(err)
	}
	if rows.Next() {
		var message types.Message
		err = rows.StructScan(&message)
		if err != nil {
			log.Fatalln(err)
		}
		rows.Close()
		return message, true
	} else {
		return types.Message{}, false
	}
}

func GetNotifications(db *sqlx.DB) []types.Notification {
	var notifications []types.Notification
	rows, err := db.Queryx("SELECT * FROM notifications")
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var Notification types.Notification
		err := rows.StructScan(&Notification)
		if err != nil {
			log.Fatalln(err)
		}
		notifications = append(notifications, Notification)
	}
	return notifications
}

func GetReports(db *sqlx.DB, schoolId string) []types.Report {
	var reports []types.Report
	rows, err := db.Queryx("SELECT * FROM reports WHERE school = $1", schoolId)
	if err != nil {
		log.Fatalln(err)
	}
	for rows.Next() {
		var report types.Report
		err := rows.StructScan(&report)
		if err != nil {
			log.Fatalln(err)
		}
		reports = append(reports, report)
	}
	return reports
}

func GetReport(db *sqlx.DB, reportId string, schoolId string) (types.Report, bool) {
	rows, err := db.Queryx("SELECT * FROM reports WHERE school = $1 AND id = $2", schoolId, reportId)
	if err != nil {
		log.Fatalln(err)
	}
	if rows.Next() {
		var report types.Report
		err = rows.StructScan(&report)
		if err != nil {
			log.Fatalln(err)
		}
		rows.Close()
		return report, true
	} else {
		return types.Report{}, false
	}
}
