package users

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	//import mysql
	"github.com/UW-Info-441-Winter-Quarter-2020/homework-ziyuguo716/servers/gateway/indexes"
	_ "github.com/go-sql-driver/mysql"
)

const sqlInsertStatement = `
INSERT INTO user (email, passhash, user_name, first_name, last_name, photo_url)
VALUES (?,?,?,?,?,?);`

const sqlUpdateStatement = `
  UPDATE user SET first_name=?, last_name=? where id=?;`

const sqlDeleteStatement = `
  Delete from user where id=?;`

const sqlGetByIDStatement = `SELECT * from user where id=?;`
const sqlGetByEmailStatement = `SELECT * from user where email=?;`
const sqlGetByUserNameStatement = `SELECT * from user where user_name=?;`

//PostgressStore stores db pointer
type PostgressStore struct {
	PostgressDB *sql.DB
	TrieNode    *indexes.TrieNode
}

//NewPostgressStore constructs a new MySQLStore
func NewPostgressStore(db *sql.DB) *PostgressStore {
	if db == nil {
		panic("nil database pointer passed to PostgressStore")
	}
	return &PostgressStore{
		PostgressDB: db,
	}
}

//FindInTrie returns relevant user IDs with names starts with prefix
func (store *PostgressStore) FindInTrie(prefix string, max int) ([]int64, error) {
	results := store.TrieNode.Find(prefix, max)
	return results, nil
}

//ScanRowsIntoUser scans rows into the user struct.
func ScanRowsIntoUser(rows *sql.Rows, err error, store *PostgressStore) ([]*User, error) {
	var users []*User
	for rows.Next() {
		newUser := &User{}
		if err := rows.Scan(&newUser.ID, &newUser.Email, &newUser.PassHash, &newUser.UserName,
			&newUser.FirstName, &newUser.LastName, &newUser.PhotoURL); err != nil {
			fmt.Printf("error scanning row: %v\n", err)
		}
		users = append(users, newUser)
	}

	//if we got an error fetching the next row, report it
	if err := rows.Err(); err != nil {
		errorStr := fmt.Sprintf("error getting next row: %v\n", err)
		return users, errors.New(errorStr)
	}
	return users, nil
}

//GetByID Get user by given ID. Will return a user struct and error (nil if no error).
func (store *PostgressStore) GetByID(id int64) (*User, error) {
	rows, err := store.PostgressDB.Query(sqlGetByIDStatement, id)
	defer rows.Close()
	if err != nil {
		return nil, errors.New("Failed GET query using user id")
	}
	newUsers, newErr := ScanRowsIntoUser(rows, err, store)
	if newErr != nil {
		return nil, errors.New("Failed scanning rows")
	}
	fmt.Printf("Get query success using user id=%d\n", id)
	return newUsers[0], nil
}

//GetByIDs Get user by given ID. Will return a user struct and error (nil if no error).
func (store *PostgressStore) GetByIDs(ids []int64) ([]*User, error) {
	if len(ids) == 0 {
		return nil, errors.New("No IDs provided")
	}
	var users []*User
	for _, id := range ids {
		rows, err := store.PostgressDB.Query(sqlGetByIDStatement, id)
		defer rows.Close()
		if err != nil {
			return nil, errors.New("Failed GET query using user id")
		}
		newUsers, newErr := ScanRowsIntoUser(rows, err, store)
		if newErr != nil {
			return nil, errors.New("Failed scanning rows")
		}
		users = append(users, newUsers[0])
	}
	return users, nil
}

//GetByEmail Get user by given email. Will return a user struct and error (nil if no error).
func (store *PostgressStore) GetByEmail(email string) (*User, error) {
	rows, err := store.PostgressDB.Query(sqlGetByEmailStatement, email)
	defer rows.Close()
	if err != nil {
		return nil, errors.New("Failed GET query using email")
	}
	newUsers, newErr := ScanRowsIntoUser(rows, err, store)
	if newErr != nil {
		return nil, errors.New("Failed scanning rows")
	}
	fmt.Printf("Get query success using email\n")
	return newUsers[0], nil

}

//GetByUserName Get user by given username. Will return a user struct and error (nil if no error).
func (store *PostgressStore) GetByUserName(username string) (*User, error) {
	rows, err := store.PostgressDB.Query(sqlGetByUserNameStatement, username)
	defer rows.Close()
	if err != nil {
		return nil, errors.New("Failed GET query using UserName")
	}
	newUsers, newErr := ScanRowsIntoUser(rows, err, store)
	if newErr != nil {
		return nil, errors.New("Failed scanning rows")
	}
	fmt.Printf("Get query success using UserName\n")
	return newUsers[0], nil

}

//Insert a contact with the given user. Will return a user struct and an error (nil if no error).
func (store *PostgressStore) Insert(user *User) (*User, error) {

	res, err := store.PostgressDB.Exec(sqlInsertStatement, user.Email, user.PassHash, user.UserName,
		user.FirstName, user.LastName, user.PhotoURL)
	if err != nil {
		fmt.Printf("error inserting new row: %v\n", err)
	} else {
		//get the auto-assigned ID for the new row
		fmt.Printf("Success inserting new row:")
		lastInsertID, err := res.LastInsertId()
		if err != nil {
			fmt.Printf("error reading new row: %v\n", err)
		}
		user.ID = lastInsertID
		AddUserToTrie(user, store)
		return user, err
	}
	return nil, errors.New("Failed Insert")

}

//Update a contact with the given ID and update fields. Will return nil or an error.
func (store *PostgressStore) Update(id int64, updates *Updates) (*User, error) {
	oldUser, err := store.GetByID(id)
	if err != nil {
		return nil, errors.New("Failed to retrieve user with given ID")
	}
	//remove user's old names from trie
	fname := strings.ToLower(oldUser.FirstName)
	lname := strings.ToLower(oldUser.LastName)
	store.TrieNode.Remove(fname, id)
	store.TrieNode.Remove(lname, id)
	//Update fields in sql
	res, err := store.PostgressDB.Exec(sqlUpdateStatement, updates.FirstName, updates.LastName, id)
	fmt.Printf("Res:%v", res)
	if err != nil {
		return nil, errors.New("Failed Update")
	}
	user, err := store.GetByID(id)
	if err != nil {
		return nil, errors.New("Failed to get user")
	}
	//Add new user names back to Trie
	AddUserToTrie(user, store)
	return user, err
}

//Delete a contact with the given ID. Will return nil or an error.
func (store *PostgressStore) Delete(id int64) error {

	res, err := store.PostgressDB.Exec(sqlDeleteStatement, id)
	if err != nil {
		return err
	}
	fmt.Printf("Res found: %v", res)
	fmt.Printf("Res deleted: %v", id)
	return nil
}

//ConnectToPostgres opens db connection
func ConnectToPostgres(dsn string) (*PostgressStore, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		fmt.Printf("error opening database: %v\n", err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		counter := 0
		for {
			time.Sleep(3 * time.Second)
			counter++
			if err = db.Ping(); err == nil {
				break
			}
			if counter == 10 {
				return nil, errors.New("Unable to ping db")
			}
		}
	}
	newPostgressStore := NewPostgressStore(db)

	return newPostgressStore, nil
}

//LoadTrie populates Trie with existing user accounts
func (store *PostgressStore) LoadTrie() error {
	store.TrieNode = indexes.NewTrieNode()
	rows, err := store.PostgressDB.Query("SELECT * from user")
	if err != nil {
		return errors.New("User store not ready")
	}
	defer rows.Close()
	newUsers, newErr := ScanRowsIntoUser(rows, err, store)
	if newErr != nil {
		return errors.New("Failed scanning rows")
	}
	for _, user := range newUsers {
		AddUserToTrie(user, store)
	}
	return nil
}

//AddUserToTrie inserts username, lastname, firstname of the given user into Trie
func AddUserToTrie(user *User, store *PostgressStore) {
	//To lower case
	fname := strings.ToLower(user.FirstName)
	lname := strings.ToLower(user.LastName)
	uname := strings.ToLower(user.UserName)
	//Split string on white-space
	fslice := strings.Fields(fname)
	lslice := strings.Fields(lname)
	uslice := strings.Fields(uname)

	for _, f := range fslice {
		store.TrieNode.Add(f, user.ID)
	}
	for _, l := range lslice {
		store.TrieNode.Add(l, user.ID)
	}
	for _, u := range uslice {
		store.TrieNode.Add(u, user.ID)
	}
}
