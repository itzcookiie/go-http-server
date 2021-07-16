package main

import (
	"encoding/json"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"io"
	"net"
	"net/http"
)

type Company struct {
	gorm.Model
	Employees []Employee
	Name, Industry string
}

type Employee struct {
	gorm.Model
	CompanyID uint
	Company   Company
	Name, Age, DateOfBirth, Job string
}

type EmployeePayload struct {
	CurrentName string
	NewName string
}

func CreateDB() (*gorm.DB, error) {
	dsn := "mayowa:abc123@tcp(localhost:3306)/golang?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	return db, err
}

func createServerHook() func(listener net.Listener) {
	return func(listener net.Listener) {
		fmt.Println("Server is live on port 8000")
	}
}

func HandleRouting(db *gorm.DB) {
	http.HandleFunc("/create", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPost {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				fmt.Println("Error!")
				return
			}
			var newEmployee *Employee
			jsonErr := json.Unmarshal(body, &newEmployee)
			print(string(body))
			if jsonErr != nil {
				fmt.Println("Error creating JSON of data")
				return
			}
			db.Create(&Employee{
				Company: Company{Name: "Marks & Spencers", Industry: "Retail"},
				Name: newEmployee.Name,
				Age: newEmployee.Age,
				DateOfBirth: newEmployee.DateOfBirth,
				Job: newEmployee.Job,
			})
			writer.Write([]byte("Created successfully"))
		}
	})
	http.HandleFunc("/update", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodPut {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				fmt.Println("Error!")
				return
			}
			var employeePayload *EmployeePayload
			err = json.Unmarshal(body, &employeePayload)
			if err != nil {
				fmt.Println("Error parsing JSON")
				return
			}
			fmt.Println(employeePayload, "temp employee")
			db.First(&Employee{}, "name = ?", employeePayload.CurrentName).Update("Name", employeePayload.NewName)
			writer.Write([]byte("Updated successfully"))
		}
	})
	http.HandleFunc("/destroy", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodDelete {
			body, err := io.ReadAll(request.Body)
			if err != nil {
				fmt.Println("Error!")
				return
			}
			var employeePayload *EmployeePayload
			jsonErr := json.Unmarshal(body, &employeePayload)
			if jsonErr != nil {
				fmt.Println("Error parsing JSON")
				return
			}
			db.Unscoped().Delete(&Employee{}, "name = ?", employeePayload.CurrentName)
			writer.Write([]byte("Deleted successfully"))
		}
	})

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		var employees []Employee
		db.First(&employees)
		json.NewEncoder(writer).Encode(employees)
	})
}

func CreateServer(hook func(listener net.Listener)) {
	l, err := net.Listen("tcp", ":8000")
	if err != nil {
		// handle error
		fmt.Println(l, "failed")
	}
	fmt.Println(l, err)
	go hook(l)
	if err := http.Serve(l, nil); err != nil {
		// handle error
	}

	fmt.Println("Finished")
}

func main() {
	db, err := CreateDB()
	if err != nil {
		fmt.Println("Could not connect to database")
		return
	}
	fmt.Println("Database ", err)
	fmt.Println("Successfully connected to ", db)

	//db.AutoMigrate(&Employee{})
	//db.AutoMigrate(&Company{})
	HandleRouting(db)
	CreateServer(createServerHook())
}
