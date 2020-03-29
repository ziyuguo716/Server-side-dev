package handlers

import (
	"time"

	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/models/users"
)

//SessionState struct: define a session state struct for this web server
//see the assignment description for the fields you should include
//remember that other packages can only see exported fields!
type SessionState struct {
	StartTime time.Time   `json:"time"`
	User      *users.User `json:"user"`
}
