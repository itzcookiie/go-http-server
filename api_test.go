package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"io"
	"net"
	"net/http"
	"testing"
)

var testEmployee = &Employee{
	Name: "Bob",
	Age: "35",
	DateOfBirth: "10th February 1990",
	Job: "Care worker",
}

func resetDB(db *gorm.DB) {
	db.Unscoped().Exec("ALTER TABLE `employees` AUTO_INCREMENT = 1;")
	db.Unscoped().Exec("ALTER TABLE `companies` AUTO_INCREMENT = 1;")
}

func createCompany(db *gorm.DB) *Company {
	marksAndSpencers := &Company{Name: "Marks & Spencers", Industry: "Retail"}
	db.Create(marksAndSpencers)
	return marksAndSpencers
}

func createTestEmployee(db *gorm.DB, company Company) {
	db.Create(&Employee{
		Name: testEmployee.Name,
		Age: testEmployee.Age,
		DateOfBirth: testEmployee.DateOfBirth,
		Job: testEmployee.Job,
		Company: company,
		})
}

func deleteTestData(db *gorm.DB, employeeName string) {
	db.Unscoped().Delete(&Employee{}, "name = ?", employeeName)
	db.Unscoped().Delete(&Company{}, "name = ?", "Marks & Spencers")
}

func getRequestHook (t *testing.T) func (listener net.Listener) {
	return func (listener net.Listener) {
		res, err := http.Get("http://localhost:8000")
		if err != nil {
			t.Errorf("GET request failed")
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		fmt.Println(string(body),2222)
		if string(body) != "hello" {
			t.Errorf("Response does not equal \"hello\"")
		}
		listener.Close()
	}
}

func createRequestHook (t *testing.T) func (listener net.Listener) {
	return func (listener net.Listener) {
		json, jsonErr := json.Marshal(*testEmployee)
		if jsonErr != nil {
			t.Errorf("%s", "Error serialising data (turning data into json format)")
		}
		data := bytes.NewReader(json)
		res, err := http.Post("http://localhost:8000/create", "application/json", data)
		if err != nil {
			t.Errorf("CREATE request failed")
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		fmt.Println(string(body),2222)
		if string(body) != "Created successfully" {
			t.Errorf("Response does not equal \"Created successfully\"")
		}
		listener.Close()
	}
}

func updateRequestHook (t *testing.T) func (listener net.Listener) {
	return func (listener net.Listener) {
		json, jsonErr := json.Marshal(&EmployeePayload{
			CurrentName: "Bob",
			NewName: "Pat",
		})
		if jsonErr != nil {
			t.Errorf("%s", "Error serialising data (turning data into json format)")
		}
		data := bytes.NewReader(json)
		client := &http.Client{}
		req, reqErr := http.NewRequest("PUT", "http://localhost:8000/update", data)
		if reqErr != nil {
			t.Errorf("UPDATE - Failed to create request object")
		}
		res, resErr := client.Do(req)
		if resErr != nil {
			t.Errorf("UPDATE - Request failed")
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("UPDATE - Failed to read response")
		}
		fmt.Println(string(body),2222)
		if string(body) != "Updated successfully" {
			t.Errorf("Response does not equal \"Updated successfully\"")
		}
		listener.Close()
	}
}

func deleteRequestHook (t *testing.T) func (listener net.Listener) {
	return func (listener net.Listener) {
		json, jsonErr := json.Marshal(&EmployeePayload{CurrentName: "Pat"})
		if jsonErr != nil {
			t.Errorf("%s", "Error serialising data (turning data into json format)")
		}
		data := bytes.NewReader(json)
		client := &http.Client{}
		req, reqErr := http.NewRequest("DELETE", "http://localhost:8000/destroy", data)
		if reqErr != nil {
			t.Errorf("DELETE - Failed to create request object")
		}
		res, resErr := client.Do(req)
		if resErr != nil {
			t.Errorf("DELETE - Request failed")
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Errorf("DELETE - Failed to read response")
		}
		fmt.Println(string(body),2222)
		if string(body) != "Deleted successfully" {
			t.Errorf("Response does not equal \"Deleted successfully\"")
		}
		listener.Close()
	}
}

func TestCRUD(t *testing.T) {
	db, err := CreateDB()
	if err != nil {
		fmt.Println("Could not connect to database")
		return
	}
	HandleRouting(db)

	t.Run("Read", func(t *testing.T) {
		fmt.Println("Start GET test")
		createCompany(db)
		CreateServer(getRequestHook(t))
		fmt.Println("Here")
		fmt.Println("Received from channel")
		deleteTestData(db, "")
		resetDB(db)
	})

	t.Run("Create", func(t *testing.T) {
		fmt.Println("Start Create test")
		createCompany(db)
		CreateServer(createRequestHook(t))
		fmt.Println("Here")
		fmt.Println("Received from channel")
		deleteTestData(db, testEmployee.Name)
		resetDB(db)
	})

	t.Run("Update", func(t *testing.T) {
		fmt.Println("Start UPDATE test")
		company := createCompany(db)
		createTestEmployee(db, *company)
		CreateServer(updateRequestHook(t))
		deleteTestData(db, "Pat")
		resetDB(db)
		fmt.Println("Finished UPDATE Test")
	})

	t.Run("Delete", func(t *testing.T) {
		fmt.Println("Start DELETE test")
		company := createCompany(db)
		createTestEmployee(db, *company)
		CreateServer(deleteRequestHook(t))
		fmt.Println("Here")
		fmt.Println("Received from channel")
		deleteTestData(db, testEmployee.Name)
		resetDB(db)
	})
}