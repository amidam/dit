package flow

import (
	"fmt"
	"strconv"
	"strings"
	"net/http"
	"text/template"

	"adrdn/dit/config"
)

const echoAllFlows		= "SELECT * FROM flow"
const newFlow			= "INSERT INTO flow (name) VALUES (?)"
const deleteFlow		= "DELETE FROM flow WHERE id = ?"
//const deleteFlowTable 	= "DROP TABLE ?"
const createTable		= "CREATE TABLE flow_"

// Flow represents the flow structure
type Flow struct {
	ID int
	Name string
}

var tmpl = template.Must(template.ParseGlob("forms/admin/flow/*"))

// ShowAllFlows displays all of the roles
func ShowAllFlows (w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	rList, err := db.Query(echoAllFlows)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	flow := Flow{}
	flowList := []Flow{}

	for rList.Next() {
		err = rList.Scan(&flow.ID, &flow.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		flowList = append(flowList, flow)
	}
	defer db.Close()
	tmpl.ExecuteTemplate(w, "Echo", flowList)
}

// New revokes the new page
func New (w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", 301)
}

// Insert adds the new flow
func Insert (w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	str := strings.Builder{}
	str.WriteString(createTable)
	var steps []string

	if r.Method == "POST" {
		name := r.FormValue("name")
		nameUpper := strings.ToUpper(name)
		numberString := r.FormValue("number")
		number, _ := strconv.Atoi(numberString)

		// define 'CTEATE TABLE flow_NAME' (
		str.WriteString(nameUpper + " (")
		
		// define coloumn names
		for i := 1; i <= number; i++ {
			steps = append(steps, "step" + strconv.Itoa(i))
		}
		
		// define query of columns with their data type
		columns := strings.Join(steps, " varchar(255), ")
		str.WriteString(columns + " varchar(255))")

		// create a table with the given name and column number
		db.Query(str.String())
		
		insForm, err := db.Prepare(newFlow)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		insForm.Exec(name)
	}
	defer db.Close()
	http.Redirect(w, r, "/admin/flow", 301)
}

// Delete removes the flow
func Delete (w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	id := r.URL.Query().Get("id")
	name := r.URL.Query().Get("name")

	_, err := db.Query("DROP TABLE " + "flow_" + strings.ToUpper(name))
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	delRow, err := db.Prepare(deleteFlow)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	delRow.Exec(id)

	defer db.Close()
	http.Redirect(w, r, "/admin/flow", 301)
}
