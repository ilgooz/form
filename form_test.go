package form

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var HTTPForm = map[string][]string{
	"friend_ids": []string{"1", "2nd"},
	"family_ids": []string{"3", "4"},
	"date":       []string{"2015-05-28T21:00:00Z"},
	"email":      []string{"ilkergoktugozturk@gmail.com"},
	"password":   []string{"12345"},
	"active":     []string{"false"},
	"colors":     []string{"blue", "red"},
}

type User struct {
	FriendIds []int64
	FamilyIds []int64
	Date      time.Time
	Name      string
	Email     string
	Password  string
	Active    bool
	Colors    []string
}

type CreateUserForm struct {
	FriendIds []int64   `form:"as:friend_ids"`
	FamilyIds []int64   `form:"as:family_ids"`
	Date      time.Time `form:"as:date"`
	Name      string    `form:"as:name,required"`
	Email     string    `form:"as:email,required"`
	Password  string    `form:"as:password,min:6,required"`
	Active    string    `form:"as:active"`
	Colors    []string  `form:"as:color"`
}

type ErrorResponse struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

func TestSomething(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		createUserForm, err := Parse(new(CreateUserForm), w, r)
		if err != nil {
			log.Fatal(err)
		}
		if createUserForm.HasError() {
			return
		}

		user := User{}
		createUserForm.ApplyTo(&user)

		date := time.Time{}
		date.UnmarshalText([]byte("2015-05-28T21:00:00Z"))

		assert.Equal(len(user.FamilyIds), 0)
		assert.Equal(user.FamilyIds, []int64{3, 4})
		assert.Equal(user.Date, date)
		assert.Equal(user.Name, "")
		assert.Equal(user.Email, "ilkergoktugozturk@gmail.com")
		assert.Equal(user.Password, "")
		assert.Equal(user.Active, true)
		assert.Equal(user.Colors, []string{"blue", "red"})

		fmt.Printf("%v", user)
	}))
	defer ts.Close()

	resp, err := http.PostForm(ts.URL, HTTPForm)
	if err != nil {
		log.Fatal(err)
	}

	errResponse := ErrorResponse{}

	if err := json.NewDecoder(resp.Body).Decode(&errResponse); err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	assert.Equal(errResponse.Message, "Unvalid Form Data")
	assert.Equal(errResponse.Fields["friend_ids"], "must be numbers")
	assert.Equal(errResponse.Fields["name"], "required")
	assert.Equal(errResponse.Fields["password"], "must be at least 6 chars long")
}
