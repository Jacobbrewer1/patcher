package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Jacobbrewer1/patcher"
	"github.com/gorilla/mux"
)

type Person struct {
	Name   string   `db:"name"`
	Age    int      `db:"age"`
	Height *float64 `db:"height"`
}

type PersonWhere struct {
	ID int `db:"id"`
}

func NewPersonWhere(id int) *PersonWhere {
	return &PersonWhere{
		ID: id,
	}
}

func (p *PersonWhere) Where() (string, []any) {
	return "id = ?", []any{p.ID}
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/people/{id}", patch).Methods(http.MethodPatch)

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}

func patch(w http.ResponseWriter, r *http.Request) {
	personIDStr := mux.Vars(r)["id"]
	personID, err := strconv.Atoi(personIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the person from the database
	person := new(Person)
	if err := json.NewDecoder(r.Body).Decode(person); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the where condition
	condition := NewPersonWhere(personID)

	// Generate the SQL
	sqlStr, args, err := patcher.GenerateSQL(
		person,
		patcher.WithTable("people"),
		patcher.WithWhere(condition),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println(sqlStr)
	fmt.Println(args)

	respStr := fmt.Sprintf("SQL:\n%s\n\nArgs:\n%v\n", sqlStr, args)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte(respStr)); err != nil {
		fmt.Println("error writing response:", err)
	}
}
