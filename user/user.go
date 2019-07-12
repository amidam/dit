package user

import (
	"net/http"
	"text/template"

	"adrdn/dit/config"
)

const echoAllUsers = "SELECT  ID, name, username FROM user"
const deleteUser = "DELETE FROM user WHERE id = ?"

// User represents the user structure
type User struct {
	ID       int
	Name     string
	Username string
	Password string
}

var tmpl = template.Must(template.ParseGlob("forms/admin/user/*"))

// DisplayAllUsers shows the list of users
func DisplayAllUsers(w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	allUsers, err := db.Query(echoAllUsers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	u := User{}
	uList := []User{}

	for allUsers.Next() {
		err = allUsers.Scan(&u.ID, &u.Name, &u.Username)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		uList = append(uList, u)
	}
	tmpl.ExecuteTemplate(w, "Echo", uList)
	defer db.Close()
}

// DeleteUser drops the user
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	id := r.URL.Query().Get("id")
	deletedUser, err := db.Prepare(deleteUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	deletedUser.Exec(id)
	defer db.Close()
	http.Redirect(w, r, "/admin/users", 301)
}