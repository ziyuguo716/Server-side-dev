package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/models/users"
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/sessions"
	"github.com/go-redis/redis"
)

func GetSessionContext() *SessionContext {

	sessionKey, sessionKeyExists := os.LookupEnv("SESSIONKEY")
	if !sessionKeyExists {
		sessionKey = "thisismykey"
	}
	redisAddr, redisAddrExists := os.LookupEnv("REDISADDR")
	if !redisAddrExists {
		redisAddr = ":6379"
	}
	dsn, dsnExists := os.LookupEnv("DSN")
	if !dsnExists {
		dsn = fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/mydb", "mypassword")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	redisSession := sessions.NewRedisStore(redisClient, 3600)

	userStore, postgressErr := users.ConnectToPostgres(dsn)
	if postgressErr != nil {
		log.Fatalf("Postgress not working")
		os.Exit(1)
	}

	handlerContext := &SessionContext{
		Key:     sessionKey,
		Session: redisSession,
		User:    userStore,
	}
	return handlerContext
}

func TestUsersHandler(t *testing.T) {
	cases := []struct {
		request        string
		contentType    string
		newUser        *users.NewUser
		expectedStatus int
	}{
		{
			"POST",
			"application/json",
			&users.NewUser{
				Email:        "gzy123@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "ziyuguo",
				FirstName:    "Ziyu",
				LastName:     "Guo",
			},
			200,
		},
	}

	for _, c := range cases {
		jsonUser, _ := json.Marshal(c.newUser)
		req, err := http.NewRequest(c.request, "/", bytes.NewBuffer(jsonUser))
		if err != nil {
			t.Fatalf(err.Error())
		}
		req.Header.Add("Content-Type", c.contentType)
		handlerContext := GetSessionContext()
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(handlerContext.UsersHandler)

		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v, wanted %v",
				status, http.StatusOK)
		}
	}
}
