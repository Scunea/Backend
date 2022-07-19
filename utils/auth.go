package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Scunea/Backend/types"
	"github.com/golang-jwt/jwt/v4"
	"github.com/jltorresm/otpgo"
	"github.com/jltorresm/otpgo/config"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/argon2"
)

func CreateToken(id string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"iat": time.Now().UnixMilli(),
		"iss": "scunea",
		"exp": time.Now().Add(time.Hour * 24 * 7).UnixMilli(),
	})

	tokenString, err := token.SignedString(CreateHmacKey())

	if err == nil {
		return tokenString
	} else {
		return ""
	}
}

func VerifyToken(tokenString string, db *sqlx.DB) bool {
	var fixedToken = strings.ReplaceAll(tokenString, "Bearer ", "")

	token, err := jwt.Parse(fixedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return CreateHmacKey(), nil
	})

	if err == nil {
		if token.Valid {
			var _, ok = GetUser(db, fixedToken, "token")
			return ok
		} else {
			return false
		}
	} else {
		return false
	}
}

func VerifySchool(tokenString string, school string, db *sqlx.DB) bool {
	var fixedToken = strings.ReplaceAll(tokenString, "Bearer ", "")
	var user, ok = GetUser(db, fixedToken, "token")
	if ok {
		var parsedSchools []string
		json.Unmarshal([]byte(user.Schools), &parsedSchools)
		return ExistsFunc(parsedSchools, func(schoolU string) bool {
			return schoolU == school
		})
	} else {
		return false
	}
}

func CreatePassword(password string) string {
	var salt = make([]byte, 64)
	rand.Reader.Read(salt)
	var hash = argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	print(b64Hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, 64*1024, 4, 32, b64Salt, b64Hash)
}

func CheckPassword(password string, hash string) bool {
	salt, _ := base64.RawStdEncoding.DecodeString(strings.Split(hash, "$")[4])
	var hashTwo = strings.Split(hash, "$")[5]
	var hashOne = argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(hashOne) == hashTwo
}

func CheckOtp(otp string, user types.User) bool {
	otpT := otpgo.TOTP{
		Algorithm: config.HmacSHA1,
		Key:       user.Tfa,
	}
	pass, err := otpT.Validate(otp)
	if err == nil {
		return pass
	} else {
		return false
	}
}

func Type(db *sqlx.DB, user types.User, schoolId string) string {
	var administers = ParseStringArray(user.Administrator)
	var teachesObj = ParseStringString(user.Teacher)

	teaches := make([]string, len(teachesObj))

	i := 0
	for k := range teachesObj {
		teaches[i] = k
		i++
	}

	var parents = GetChildren(db, user)
	if ExistsFunc(administers, func(school string) bool {
		return school == schoolId
	}) {
		return "administrator"
	} else if ExistsFunc(teaches, func(school string) bool {
		return school == schoolId
	}) {
		return "teacher"
	} else if len(parents) > 0 {
		return "parent"
	} else {
		return "student"
	}
}

func GetAvailable(db *sqlx.DB, myUser types.User, schoolId string) []types.SimpleUser {
	var preAvailable = GetUsers(db)

	for i, user := range preAvailable {
		if !ExistsFunc(ParseStringArray(user.Schools), func(school string) bool {
			return school == schoolId
		}) || (Type(db, user, schoolId) == "student" && Type(db, myUser, schoolId) != "teacher" && Type(db, myUser, schoolId) != "administrator") {
			preAvailable = append(preAvailable[:i], preAvailable[i+1:]...)
		}
	}

	var available []types.SimpleUser

	for _, user := range preAvailable {
		var parsedChildren = GetChildren(db, user)

		var children []string

		for _, child := range parsedChildren {
			children = append(children, child.Id)
		}

		available = append(available, types.SimpleUser{
			Id:       user.Id,
			Name:     user.Name,
			Teacher:  ParseStringString(user.Teacher)[schoolId],
			Children: children,
			Type:     Type(db, user, schoolId),
		})
	}

	return available
}
