package request

import (
	"fmt"
	"strings"
	"strconv"
	"net/http"
	"text/template"

	"adrdn/dit/flow"
	"adrdn/dit/config"
)

const echoAllFlow = "SELECT * FROM flow"
const echoALLRequest = "SELECT * FROM request"
const echoOneRequest = "SELECT ID, type, current_step, description FROM request WHERE ID = ?"
const addNewRequest = "INSERT INTO request(type, current_step, description) VALUES(?, ?, ?)"
const updateRequest = "UPDATE request SET current_step = ?, description = ? WHERE ID = ?"

var tmpl = template.Must(template.ParseGlob("forms/request/*"))

// Request represents the request structure
type Request struct {
	ID	 		int
	Type 		string
	PriorStep	string
	CurrentStep int
	NextStep	string
	Termination int
	Completion	int
	Deletion	int
	Status		string
	Description	string
	IsFirstStep	bool
	IsLastStep	bool
}

// New starts a new request of the select flow
func New(w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	flowEntity := flow.Flow{}
	flowList := []flow.Flow{}
	
	flows, err := db.Query(echoAllFlow)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	for flows.Next() {
		err = flows.Scan(&flowEntity.ID, &flowEntity.Name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		flowList = append(flowList, flowEntity)
	}
	db.Close()
	tmpl.ExecuteTemplate(w, "New", flowList)
}

// Insert adds the new entity
func Insert(w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	if r.Method == "POST" {
		selectedFlow := r.FormValue("flow")
		description := r.FormValue("description")

		newRequest, err :=db.Prepare(addNewRequest)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		// current_step is assinged to 2 because at this stage the request is already created
		_, err = newRequest.Exec(selectedFlow, 2, description)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	db.Close()
	http.Redirect(w, r, "/request/view", 301)
}

// Echo displays all of the requests
func Echo(w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	req := Request{}
	reqList := []Request{}

	requests, err := db.Query(echoALLRequest)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	for requests.Next() {
		err = requests.Scan(&req.ID, &req.Type, &req.CurrentStep, &req.Termination, &req.Completion, &req.Deletion, &req.Description)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		if req.Termination == 0 && req.Completion == 0 {
			req.Status = "In Process"
		} else if req.Termination != 0 {
			req.Status = "Terminated"	
		} else {
			req.Status = "Completed"
		}
		reqList = append(reqList, req)
	}
	db.Close()
	tmpl.ExecuteTemplate(w, "Echo", reqList)
}

// ShowDetails revoke the detail page
func ShowDetails(w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()
	req := Request{}

	ID := r.URL.Query().Get("id")

	request, err := db.Query(echoOneRequest, ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	for request.Next() {
		err = request.Scan(&req.ID, &req.Type, &req.CurrentStep, &req.Description)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// If IsFirstStep is true, then show the terminate option on the Detail page
	if req.CurrentStep == 1 {
		req.IsFirstStep = true
	}

	intPreStep := req.CurrentStep - 1
	stringPreStep := strconv.Itoa(intPreStep)
	if intPreStep == 0 {
		req.PriorStep = "You are the creator of this request"
	} else {
		// Fetch string value of the previous Step
		preStep, err := db.Query("SELECT step" + stringPreStep + " FROM flow_" + strings.ToUpper(req.Type))
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		for preStep.Next() {
			err = preStep.Scan(&req.PriorStep)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	}

	intNextStep := req.CurrentStep + 1
	stringNextStep := strconv.Itoa(intNextStep)

	// Fetch string value of the next Step
	nextStep, err := db.Query("SELECT step" + stringNextStep + " FROM flow_" + strings.ToUpper(req.Type))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	for nextStep.Next() {
		err = nextStep.Scan(&req.NextStep)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	db.Close()
	tmpl.ExecuteTemplate(w, "Detail", req)
}

// Update changes the request based on user decision
func Update(w http.ResponseWriter, r *http.Request) {
	db := config.DbConn()

	if r.Method == "POST" {
		ID := r.URL.Query().Get("id")
		currentStep := r.URL.Query().Get("cs")
		intCurrentStep, _ := strconv.Atoi(currentStep)
		description := r.FormValue("description")
		decision := r.FormValue("decision")
		if decision == "approve" {
			intCurrentStep++
			currentStep = strconv.Itoa(intCurrentStep)
		} else if decision == "reject" {
			intCurrentStep--
			currentStep = strconv.Itoa(intCurrentStep)
		} else {

		}
		request, err := db.Prepare(updateRequest)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		_, err = request.Exec(currentStep, description, ID)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	defer db.Close()
	http.Redirect(w, r, "/request/view", 301)
}