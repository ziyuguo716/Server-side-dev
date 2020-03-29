package users

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostgressStore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("error creating sqlmock: %v", err)
	}
	defer db.Close()
	postgresstore := NewPostgressStore(db)

	//Insert()
	expectedInsertSQL := regexp.QuoteMeta(sqlInsertStatement)
	// expectedUpdateSQL := regexp.QuoteMeta(sqlUpdateStatement)
	// expectedGetSQL := regexp.QuoteMeta(sqlGetByIDStatement)
	var newID int64 = 1
	newUser := User{
		Email:     "gzy@uw.edu",
		PassHash:  []byte{3, 4, 1},
		UserName:  "gzy",
		FirstName: "first",
		LastName:  "last",
		PhotoURL:  "sssss",
	}
	//columns := []string{"id", "email", "passhash", "user_name", "first_name", "last_name", "photo_url"}
	//INSERT
	mock.ExpectExec(expectedInsertSQL).
		// with these arguments
		WithArgs(
			newUser.Email,
			newUser.PassHash,
			newUser.UserName,
			newUser.FirstName,
			newUser.LastName,
			newUser.PhotoURL,
		).
		// with these results
		WillReturnResult(sqlmock.NewResult(newID, 1))
	// UPDATE
	// mock.ExpectExec(expectedUpdateSQL).
	// 	WithArgs("firstname", "last", 1).
	// 	WillReturnResult(sqlmock.NewResult(newID, 1))

	// mock.ExpectQuery(expectedGetSQL).
	// 	WithArgs(newID).
	// 	WillReturnRows(sqlmock.NewRows(columns).AddRow(newID, newUser.Email, newUser.PassHash, newUser.UserName, newUser.FirstName, newUser.LastName, newUser.PhotoURL))
	// //DELETE
	// mock.ExpectExec("DELETE FROM contacts").
	// 	WithArgs(sql.Named("id", 1)).
	// 	WillDelayFor(time.Second).
	// 	WillReturnResult(sqlmock.NewResult(1, 1))

	// if err := mock.ExpectationsWereMet(); err != nil {
	// 	t.Errorf("there were unfulfilled expectations: %s", err)
	// }

	//Insert()
	insertedUser, err := postgresstore.Insert(&newUser)
	if err != nil {
		t.Fatalf("unexpected error during successful insert: %v", err)
	}
	if insertedUser == nil {
		t.Fatal("nil user returned from insert")
	} else if insertedUser.ID != newID {
		t.Fatalf("incorrect new ID: expected %d but got %d", newID, insertedUser.ID)
	}

	// newUpdates := Updates{
	// 	FirstName: "firstname",
	// 	LastName:  "last",
	// }
	// //Update
	// updatedUser, err := postgresstore.Update(newID, &newUpdates)
	// if err != nil {
	// 	t.Fatalf("unexpected error during successful update: %v", err)
	// }
	// if updatedUser == nil {
	// 	t.Fatalf("nil user returned from update")
	// } else if updatedUser.ID != newID {
	// 	t.Fatalf("incorrect new iD: expected %d but got %d", newID, updatedUser.ID)
	// }

}

// func TestSelectById(t *testing.T) {

// }

// func TestSelectByEmail(t *testing.T) {

// }

// func TestSelectByUserName(t *testing.T) {

// }

// func TestInsert(t *testing.T) {

// }

// func TestUpdate(t *testing.T) {}

// func TestDelete(t *testing.T) {}
