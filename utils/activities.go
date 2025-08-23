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
	"encoding/json"
	"strconv"

	"github.com/Scunea/Backend/types"
	"github.com/jmoiron/sqlx"
)

func ParseActivities(db *sqlx.DB, activities []types.Activity) []types.ActivityParsed {
	var parsedActivities = []types.ActivityParsed{}

	for _, activity := range activities {
		var activityParsed types.ActivityParsed
		activityParsed.Id = activity.Id
		activityParsed.Title = activity.Title
		activityParsed.Description = activity.Description

		var files = []types.IdAndName{}
		json.Unmarshal([]byte(activity.Files), &files)
		activityParsed.Files = files

		activityParsed.Type = activity.Type
		activityParsed.Delivery = activity.Delivery
		user, ok := GetUser(db, activity.Author, "id")
		if ok {
			activityParsed.Author = types.IdAndName{
				Id:   activity.Author,
				Name: user.Name,
			}
		}
		date, err := strconv.ParseInt(activity.Date, 0, 64)
		if err == nil {
			activityParsed.Date = date
		}
		expiration, err := strconv.ParseInt(activity.Expiration, 0, 64)
		if err == nil {
			activityParsed.Expiration = expiration
		}
		activityParsed.Receiver = ParseStringArray(activity.Receiver)

		var delivered map[string]types.Deliver
		json.Unmarshal([]byte(activity.Delivered), &delivered)
		activityParsed.Delivered = delivered

		activityParsed.Result = ParseStringString(activity.Result)
		activityParsed.Viewed = ParseStringArray(activity.Viewed)
		activityParsed.School = activity.School
		parsedActivities = append(parsedActivities, activityParsed)
	}
	return parsedActivities
}

func AlonifyActivities(parsedActivities []types.ActivityParsed, userId string) []types.ActivityParsedAlone {
	var parsedActivitiesAlone = []types.ActivityParsedAlone{}
	for _, activity := range parsedActivities {
		var activityParsedAlone types.ActivityParsedAlone
		activityParsedAlone.Id = activity.Id
		activityParsedAlone.Title = activity.Title
		activityParsedAlone.Description = activity.Description
		activityParsedAlone.Files = activity.Files
		activityParsedAlone.Type = activity.Type
		activityParsedAlone.Delivery = activity.Delivery
		activityParsedAlone.Author = activity.Author
		activityParsedAlone.Date = activity.Date
		activityParsedAlone.Expiration = activity.Expiration
		activityParsedAlone.Receiver = activity.Receiver
		activityParsedAlone.Delivered = activity.Delivered[userId]
		activityParsedAlone.Result = activity.Result[userId]
		activityParsedAlone.Viewed = ExistsFunc(activity.Viewed, func(user string) bool {
			return userId == user
		})
		activityParsedAlone.School = activity.School
		parsedActivitiesAlone = append(parsedActivitiesAlone, activityParsedAlone)
	}

	return parsedActivitiesAlone
}
