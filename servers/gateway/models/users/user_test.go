package users

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

//TODO: add tests for the various functions in user.go, as described in the assignment.
//use `go test -cover` to ensure that you are covering all or nearly all of your code paths.

func TestValidate(t *testing.T) {
	cases := []struct {
		name           string
		hint           string
		newUser        *NewUser
		expectedOutput error
	}{
		{
			"All valid inputs",
			"Make sure all inputs are valid",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			nil,
		},
		{
			"Invalid email address",
			"See mail.ParseAddress",
			&NewUser{
				Email:        "bchong_AT_uw_DOT_edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			fmt.Errorf("Invalid email address"),
		},
		{
			"Password shorter than 6 characters",
			"Make sure to check password length",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "pass",
				PasswordConf: "pass",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			fmt.Errorf("Password cannot be less than 6 characters"),
		},
		{
			"Password and PasswordConf don't match",
			"Make sure that password and password confirmation match",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "not_password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			fmt.Errorf("Password and password confirmation do not match"),
		},
		{
			"UserName is empty",
			"Make sure that username length > 0",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			fmt.Errorf("Username cannot be empty"),
		},
		{
			"UserName has spaces",
			"Make sure username doesn't have spaces in it",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     " bch ong ",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			fmt.Errorf("Username cannot have spaces"),
		},
	}

	for _, c := range cases {
		err := c.newUser.Validate()
		if err != nil && err != io.EOF {
			if !reflect.DeepEqual(err, c.expectedOutput) {
				t.Errorf("case %s: Incorrect result:\nEXPECTED: %s\nACTUAL: %s\nHINT: %s\n",
					c.name, c.expectedOutput, err, c.hint)
			}
		}
	}
}

func TestToUser(t *testing.T) {
	cases := []struct {
		name                string
		hint                string
		newUser             *NewUser
		expectedErrorOutput error
	}{
		{
			"All valid inputs",
			"Make sure all inputs are valid",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			nil,
		},
		{
			"Invalid inputs",
			"Make sure Validate() returns err",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			fmt.Errorf("Username cannot be empty"),
		},
		{
			"Email address with mixed case",
			"Make sure to lowercase/uppercase the entire email",
			&NewUser{
				Email:        "bChOnG@uw.eDu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			nil,
		},
		{
			"Email address with leading and trailing whitespace",
			"Make sure to trim the email",
			&NewUser{
				Email:        " bchong@uw.edu ",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			nil,
		},
		{
			"Password is hashed",
			"Make sure that SetPassword() is implmented",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			nil,
		},
		{
			"Gravatar url is correct",
			"Make sure to hash the email",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			nil,
		},
	}

	for _, c := range cases {
		user, err := c.newUser.ToUser()
		if err != nil && err != io.EOF {
			if !reflect.DeepEqual(err, c.expectedErrorOutput) {
				t.Errorf("case %s: Incorrect result:\nEXPECTED: %s\nACTUAL: %s\nHINT: %s\n",
					c.name, c.expectedErrorOutput, err, c.hint)
			}
		}
		switch c.name {
		case "Email address with mixed case":
			if !reflect.DeepEqual(user.Email, strings.ToLower(c.newUser.Email)) {
				t.Errorf("case %s: Incorrect result: %s\n %s", c.name, "Email was not cleaned.", c.hint)
			}
		case "Email address with leading and trailing whitespace":
			if !reflect.DeepEqual(user.Email, strings.TrimSpace(c.newUser.Email)) {
				t.Errorf("case %s: Incorrect result: %s\n %s", c.name, "Email was not cleaned.", c.hint)
			}
		case "Password is hashed":
			err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(c.newUser.Password))
			if err != nil {
				t.Errorf("case %s: Incorrect result: %s\n HINT: %s", c.name, "Password hash is incorrect.", c.hint)
			}
		case "Gravatar url is correct":
			hash := md5.Sum([]byte(c.newUser.Email))
			md5Email := "https://www.gravatar.com/avatar/" + string(hash[:])
			if user.PhotoURL != md5Email {
				t.Errorf("case %s: Incorrect result: %s", c.name, "Gravatar URL is incorrect.")
			}
		default:
			// if err != nil {
			// 	expectedJSON, _ := json.MarshalIndent(c.expectedErrorOutput, "", "  ")
			// 	actualJSON, _ := json.MarshalIndent(err, "", "  ")
			// 	t.Errorf("case %s: Incorrect result:\nEXPECTED: %s\nACTUAL: %s\nHINT: %s\n",
			// 		c.name, string(expectedJSON), string(actualJSON), c.hint)
			// }
		}
	}
}

func TestFullName(t *testing.T) {
	cases := []struct {
		name           string
		hint           string
		newUser        *NewUser
		expectedOutput string
	}{
		{
			"All valid inputs",
			"Make sure all inputs are valid",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			"Brandon Chong",
		},
		{
			"Empty last name",
			"Make sure to check for empty string in names",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "",
			},
			"Brandon",
		},
		{
			"Last name only has spaces",
			"Make sure to trim the names",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "     ",
			},
			"Brandon",
		},
		{
			"Empty first and last names",
			"Make sure to trim the names",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "",
				LastName:     "",
			},
			"",
		},
	}

	for _, c := range cases {
		user, err := c.newUser.ToUser()
		if err != nil && err != io.EOF {
			t.Errorf("case %s: unexpected error %v\nHINT: %s\n", c.name, err, c.hint)
		}
		fullName := user.FullName()
		if !reflect.DeepEqual(fullName, c.expectedOutput) {
			expectedString := c.expectedOutput
			actualString := fullName
			t.Errorf("case %s: Incorrect result:\nEXPECTED: %s\nACTUAL: %s\nHINT: %s\n",
				c.name, expectedString, actualString, c.hint)
		}
	}
}

func TestAuthenticate(t *testing.T) {
	cases := []struct {
		name           string
		hint           string
		newUser        *NewUser
		testPassword   string
		expectedOutput error
	}{
		{
			"All valid inputs",
			"Make sure all inputs are valid",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			"password",
			nil,
		},
		{
			"Incorrect password",
			"Make sure to use bcrypt functions",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			"not_password",
			fmt.Errorf("Wrong Password"),
		},
		{
			"Empty password",
			"Make sure to use bcrypt functions",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			"",
			fmt.Errorf("Wrong Password"),
		},
	}

	for _, c := range cases {
		user, err := c.newUser.ToUser()
		if err != nil && err != io.EOF {
			t.Errorf("case %s: unexpected error %v\nHINT: %s\n", c.name, err, c.hint)
		}
		authErr := user.Authenticate(c.testPassword)
		if !reflect.DeepEqual(authErr, c.expectedOutput) {
			t.Errorf("case %s: Incorrect result:\nEXPECTED: %s\nACTUAL: %s\nHINT: %s\n",
				c.name, authErr, c.expectedOutput, c.hint)
		}
	}
}

func TestApplyUpdates(t *testing.T) {
	cases := []struct {
		name           string
		hint           string
		newUser        *NewUser
		updates        *Updates
		expectedOutput error
	}{
		{
			"All valid inputs",
			"Make sure all inputs are valid",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			&Updates{
				FirstName: "NotBrandon",
				LastName:  "NotChong",
			},
			nil,
		},
		{
			"Spaces in names",
			"Make sure to trim names",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			&Updates{
				FirstName: " NotBrandon ",
				LastName:  " NotChong ",
			},
			nil,
		},
		{
			"Empty names",
			"If both FirstName and LastName are empty, throw error",
			&NewUser{
				Email:        "bchong@uw.edu",
				Password:     "password",
				PasswordConf: "password",
				UserName:     "bchong",
				FirstName:    "Brandon",
				LastName:     "Chong",
			},
			&Updates{
				FirstName: "",
				LastName:  "",
			},
			fmt.Errorf("First and Last Name cannot be empty at the same time"),
		},
	}

	for _, c := range cases {
		user, err := c.newUser.ToUser()
		if err != nil && err != io.EOF {
			t.Errorf("case %s: unexpected error %v\nHINT: %s\n", c.name, err, c.hint)
		}
		updateErr := user.ApplyUpdates(c.updates)
		if !reflect.DeepEqual(updateErr, c.expectedOutput) {
			expectedJSON, _ := json.MarshalIndent(c.expectedOutput, "", "  ")
			actualJSON, _ := json.MarshalIndent(updateErr, "", "  ")
			t.Errorf("case %s: Incorrect result:\nEXPECTED: %s\nACTUAL: %s\nHINT: %s\n",
				c.name, string(expectedJSON), string(actualJSON), c.hint)
		}
	}
}
