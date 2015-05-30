package form

import (
	"encoding/json"
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
	"point":      []string{"12.09"},
}

type CreateUserForm struct {
	FriendIds []int64   `form:"as:friend_ids"`
	FamilyIds []int64   `form:"as:family_ids"`
	Date      time.Time `form:"as:date"`
	Name      string    `form:"as:name,required"`
	Email     string    `form:"as:email,required"`
	Password  string    `form:"as:password,min:6,required"`
	Active    bool      `form:"as:active"`
	Colors    []string  `form:"as:colors"`
	Point     float32   `form:"as:point"`
}

type ErrorResponse struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

func TestSomething(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	assert := assert.New(t)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := CreateUserForm{}

		createUserForm, err := Parse(&u, w, r)
		if err != nil {
			log.Fatal(err)
		}

		assert.Equal(createUserForm.HasError(), true)

		date := time.Time{}
		date.UnmarshalText([]byte("2015-05-28T21:00:00Z"))

		assert.Equal(len(u.FriendIds), 0)
		assert.Equal(u.FamilyIds, []int64{3, 4})
		assert.Equal(u.Date, date)
		assert.Equal(u.Name, "")
		assert.Equal(u.Email, "ilkergoktugozturk@gmail.com")
		assert.Equal(u.Password, "")
		assert.Equal(u.Active, false)
		assert.Equal(u.Colors, []string{"blue", "red"})
		assert.Equal(u.Point, float32(12.09))
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
