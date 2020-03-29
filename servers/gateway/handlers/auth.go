package handlers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/sessions"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/models/users"
)

//TODO: define HTTP handler functions as described in the
//assignment description. Remember to use your handler context
//struct as the receiver on these functions so that you have
//access to things like the session store and user store.

const maxNumResult = 20

// UsersHandler handles all requests for the users resource
func (context *SessionContext) UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		context.SearchUserHandler(w, r)
	} else if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-type")
		if contentType == "application/json" {
			decoder := json.NewDecoder(r.Body)
			var user users.NewUser
			err := decoder.Decode(&user)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to decode"))
			} else {
				user, err := user.ToUser()
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(err.Error()))
					return
				}
				//Create a new session state based on the current time
				newState := &SessionState{
					StartTime: time.Now(),
					User:      user,
				}
				//Insert decoded user into db
				insUser, insErr := context.User.Insert(user)
				if insErr != nil {
					w.Write([]byte("Error Inserting into Database"))
					return
				}
				//begin a new session based on the session context
				_, sessionErr := sessions.BeginSession(context.Key, context.Session, newState, w)
				if sessionErr != nil {
					w.WriteHeader(http.StatusUnauthorized)
					w.Write([]byte("Error Beginning a new session"))
					return
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				//Encode user. HashPass and Email already defined as hidden in User struct
				json.NewEncoder(w).Encode(insUser)
			}
		} else {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte("The request body must be in JSON format"))
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

//SearchUserHandler handles user search requests by finding results from Trie
func (context *SessionContext) SearchUserHandler(w http.ResponseWriter, r *http.Request) {
	//Check authorization
	sessionState := &SessionState{}
	_, err := sessions.GetState(r, context.Key, context.Session, sessionState)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized session"))
		return
	}

	//Get query param from api call
	query := r.FormValue("q")
	if strings.TrimSpace(query) == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Search query cannot be empty"))
		return
	}
	//Search for top 20 userID whose name started with query prefix
	results, _ := context.User.FindInTrie(query, maxNumResult)
	//Retrive user info from db
	users, err := context.User.GetByIDs(results)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Unable to find users"))
		return
	}
	//sort users by username
	sort.Slice(users, func(i, j int) bool {
		return users[i].UserName < users[j].UserName
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

//SpecificUserHandler handles request from a specific user with UserID
func (context *SessionContext) SpecificUserHandler(w http.ResponseWriter, r *http.Request) {
	sessionState := &SessionState{}
	_, err := sessions.GetState(r, context.Key, context.Session, sessionState)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized session"))
		return
	}
	user := sessionState.User

	//u is the user param passed in by client
	_, u := filepath.Split(r.URL.Path)
	println(u)
	if r.Method == http.MethodGet {
		if u == "me" {
			meUser := user
			w.Header().Set("Content-Type", "application/json")
			// w.WriteHeader(http.StatusOK)
			encoded, err := json.Marshal(meUser)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to encode users"))
				return
			}
			w.Write(encoded)
			return
		}
		//convert string to int64 for comparison
		uid, _ := strconv.ParseInt(u, 10, 64)

		requestedUser, err := context.User.GetByID(uid)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("User not found!"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(requestedUser)
		return
	} else if r.Method == http.MethodPatch {
		//convert string to int64 for comparison
		uid, _ := strconv.ParseInt(u, 10, 64)

		if u != "me" && uid != user.ID {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("User ID invalid!"))
			return
		}
		contentType := r.Header.Get("Content-type")
		if contentType != "application/json" {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			w.Write([]byte("The request body must be in JSON!"))
			return
		}
		var update users.Updates
		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to decode JSON"))
			return
		}
		errUpdate := user.ApplyUpdates(&update)
		if errUpdate != nil {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("FirstName or LastName cannot be empty"))
			return
		}
		updatedUser, err := context.User.Update(user.ID, &update)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("User could not be updated."))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		ec := json.NewEncoder(w)
		err = ec.Encode(updatedUser)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Error encoding"))
			return
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// SessionsHandler handles sessions by allowing clients to being a new session
// using their user credentials.
func (context *SessionContext) SessionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" {
			decoder := json.NewDecoder(r.Body)
			var cred users.Credentials
			err := decoder.Decode(&cred)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to decode JSON"))
				return
			}
			//Get user from email in credential
			user, userError := context.User.GetByEmail(cred.Email)
			if userError != nil {
				time.Sleep(1 * time.Second)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("invalid credentials"))
				return
			}
			//Authenticate user with pwd in credential
			authErr := user.Authenticate(cred.Password)
			if authErr != nil {
				time.Sleep(1 * time.Second)
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Invalid credentials"))
				return
			}
			//Create a new session state based on the current time
			newState := &SessionState{
				StartTime: time.Now(),
				User:      user,
			}
			//Begin a new session based on the session context
			_, sessionErr := sessions.BeginSession(context.Key, context.Session, newState, w)
			if sessionErr != nil {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Error Beginning a new session"))
				return
			}
			//Encode user. HashPass and Email already defined as hidden in User struct

			//respond back to client with encoded user
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)

			return
		}
		w.WriteHeader(http.StatusUnsupportedMediaType)
		w.Write([]byte("The request body must be in JSON!"))
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// SpecificSessionHandler handles closing a specific authenticated sessions.
func (context *SessionContext) SpecificSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		_, identifier := filepath.Split(r.URL.Path)
		if identifier != "mine" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Wrong session to be closed!"))
			return
		}
		sessions.EndSession(r, context.Key, context.Session)
		w.Write([]byte("Signed out"))
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
