package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"rifluxyss/database"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Employee struct {
	ID          primitive.ObjectID `json:"id" form:"id," bson:"_id,omitempty"`
	Email       string             `json:"email,omitempty" bson:"email,omitempty"`
	FirstName   string             `json:"first_name,omitempty" bson:"first_name,omitempty"`
	LastName    string             `json:"last_name,omitempty" bson:"last_name,omitempty"`
	PhoneNumber string             `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
}

func main() {
	route := mux.NewRouter()

	route.HandleFunc("/api/employee", createemployee).Methods("POST")
	route.HandleFunc("/api/employee", getemployees).Methods("GET")
	route.HandleFunc("/api/employee/{id}", updateStudent).Methods("PUT")
	route.HandleFunc("/api/upload/employee", EmployeeUpload).Methods("POST")
	route.HandleFunc("/api/employeelist", EmployeeList).Methods("POST")

	log.Fatal(http.ListenAndServe(":8090", route))
}
func createemployee(W http.ResponseWriter, r *http.Request) {

	W.Header().Add("Content-Type", "application/json")

	var employee Employee

	err := json.NewDecoder(r.Body).Decode(&employee)
	if err != nil {
		W.WriteHeader(400)
		json.NewEncoder(W).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}
	duplicateCheck, err := EmployeeDuplocateCheck(&employee)
	if duplicateCheck != nil {
		W.WriteHeader(400)
		json.NewEncoder(W).Encode(map[string]string{"error": "Employee Duplicate Founded"})
		return
	}
	if err != nil {
		W.WriteHeader(500)
		json.NewEncoder(W).Encode(map[string]string{"error": err.Error()})
		return
	}
	collection := database.ConnectDB()
	// insert our book model.
	result, err := collection.InsertOne(context.TODO(), employee)

	if err != nil {
		W.WriteHeader(500)
		json.NewEncoder(W).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(W).Encode(result)
}
func getemployees(W http.ResponseWriter, r *http.Request) {
	W.Header().Set("Content-Type", "application/json")

	// we created Book array
	var employee []Employee

	// bson.M{},  we passed empty filter. So we want to get all data.
	collection := database.ConnectDB()
	cur, err := collection.Find(context.TODO(), bson.M{})

	if err != nil {
		W.WriteHeader(500)
		json.NewEncoder(W).Encode(map[string]string{"error": err.Error()})
		return
	}

	defer cur.Close(context.TODO())

	for cur.Next(context.TODO()) {

		var employees Employee

		err := cur.Decode(&employees)
		if err != nil {
			W.WriteHeader(500)
			json.NewEncoder(W).Encode(map[string]string{"error": err.Error()})
			return
		}

		employee = append(employee, employees)
	}

	if err := cur.Err(); err != nil {
		W.WriteHeader(500)
		json.NewEncoder(W).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(W).Encode(employee)
}
func updateStudent(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var student Employee
	err := json.NewDecoder(r.Body).Decode(&student)
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request payload"})
		return
	}
	update := bson.M{
		"$set": student,
	}
	var Params = mux.Vars(r)
	objID, _ := primitive.ObjectIDFromHex(Params["id"])
	collection := database.ConnectDB()
	updateResult, err := collection.UpdateMany(context.Background(), bson.M{"_id": objID}, update)

	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	json.NewEncoder(w).Encode(updateResult)
}
func EmployeeDuplocateCheck(employee *Employee) (*Employee, error) {
	mainPipeline := []bson.M{}
	query := []bson.M{}
	if employee.Email != "" {
		query = append(query, bson.M{"email": employee.Email})
	}
	if employee.FirstName != "" {
		query = append(query, bson.M{"first_name": employee.FirstName})
	}
	if employee.LastName != "" {
		query = append(query, bson.M{"last_name": employee.LastName})
	}
	if employee.PhoneNumber != "" {
		query = append(query, bson.M{"phone_number": employee.PhoneNumber})
	}
	if len(query) > 0 {
		mainPipeline = append(mainPipeline, bson.M{"$match": bson.M{"$and": query}})
	}
	b, _ := json.Marshal(mainPipeline)
	fmt.Println("ddd==>", string(b))
	collection := database.ConnectDB()
	cursor, err := collection.Aggregate(context.TODO(), mainPipeline, nil)
	if err != nil {
		return nil, err
	}
	var Employees []Employee
	var Employee *Employee

	if err = cursor.All(context.TODO(), &Employees); err != nil {
		return nil, err
	}
	if len(Employees) > 0 {
		Employee = &Employees[0]
	}
	return Employee, nil
}
func EmployeeUpload(w http.ResponseWriter, r *http.Request) {
	// platform := r.URL.Query().Get("platform")
	file, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		// fmt.Fprintf(w, err.Error())
		return
	}
	defer file.Close()
	defer r.Body.Close()
	const (
		MAXCOLUMN       = 4
		OMITROWS        = 0
		EMAILCOLUMN     = 0
		FIRSTNAMECOLUMN = 1
		LASTNAMECOLUMN  = 2
		MOBILENOCOLUMN  = 3
	)
	employees := make([]Employee, 0)

	fmt.Println("started reading file")
	f, err := excelize.OpenReader(file)
	if err != nil {
		w.WriteHeader(422)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	rows := f.GetRows("Sheet1")
	//var errors []string
	fmt.Println("started looping")
	for rowIndex, row := range rows {
		fmt.Println("row no === ", rowIndex)
		if rowIndex <= OMITROWS {
			continue
		}
		if len(row) < MAXCOLUMN {
			w.WriteHeader(422)
			err := errors.New("Excel is not upto the format")
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		employee := new(Employee)
		employee.Email = row[EMAILCOLUMN]
		employee.FirstName = row[FIRSTNAMECOLUMN]
		employee.LastName = row[LASTNAMECOLUMN]
		employee.PhoneNumber = row[MOBILENOCOLUMN]
		err = SaveEmployee(*employee)
		if err != nil {
			w.WriteHeader(422)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return

		}
		employees = append(employees, *employee)
	}
	fmt.Println("no.of.employee==>", len(employees))

	json.NewEncoder(w).Encode(employees)
}
func SaveEmployee(employee Employee) error {
	// insert our book model.
	duplicateCheck, err := EmployeeDuplocateCheck(&employee)
	if duplicateCheck != nil {
		fmt.Println("duplicateCheck", duplicateCheck)
		return errors.New("Employee Duplicate Founded")
	}
	if err != nil {
		return err
	}
	collection := database.ConnectDB()
	result, err := collection.InsertOne(context.TODO(), employee)
	if err != nil {
		return err
	}
	fmt.Println("result", result)

	return err
}
func EmployeeList(w http.ResponseWriter, r *http.Request) {
	var employee []Employee
	collection := database.ConnectDB()
	cur, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var employees Employee
		err := cur.Decode(&employees)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		employee = append(employee, employees)
	}
	excel := excelize.NewFile()
	sheet1 := "EmployeeList"
	index := excel.NewSheet(sheet1)
	rowNo := 1
	excel.SetActiveSheet(index)
	excel.MergeCell(sheet1, "A1", "E1")
	style1, err := excel.NewStyle(`{"fill":{"type":"pattern","color":["#FFDC6D"],"pattern":1},"alignment":{"horizontal":"center","vertical":"center"},"font":{"bold":true}}`)
	if err != nil {
		fmt.Println(err)
	}
	excel.SetCellStyle(sheet1, fmt.Sprintf("%v%v", "A", rowNo), fmt.Sprintf("%v%v", "E", rowNo), style1)
	excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "C", rowNo), fmt.Sprintf("%v-%v", sheet1, "EmployeeList"))
	rowNo++
	excel.SetCellStyle(sheet1, fmt.Sprintf("%v%v", "A", rowNo), fmt.Sprintf("%v%v", "E", rowNo), style1)
	excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "A", rowNo), "S.No")
	excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "B", rowNo), "Email")
	excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "C", rowNo), "First Name")
	excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "D", rowNo), "Last Name")
	excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "E", rowNo), "Phone Number")
	rowNo++
	for k, v := range employee {
		excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "A", rowNo), k+1)
		excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "B", rowNo), v.Email)
		excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "C", rowNo), v.FirstName)
		excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "D", rowNo), v.LastName)
		excel.SetCellValue(sheet1, fmt.Sprintf("%v%v", "E", rowNo), v.PhoneNumber)
		rowNo++
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=EmployeeList.xlsx")
	w.Header().Set("ocntent-Transfer-Encoding", "binary")
	excel.Write(w)
}
