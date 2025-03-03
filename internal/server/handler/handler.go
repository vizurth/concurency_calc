package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/vizurth/concurency_calc/pkg/parser"
	"net/http"
	"strconv"
	"sync"
)

var taskMutex sync.Mutex

type Expression struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	Result string `json:"result"`
}
type GetExpressionsResponse struct {
	Expressions []Expression `json:"expressions"`
}
type GetExpressionByIdResponse struct {
	Expression Expression `json:"expression"`
}

type GetTaskResponse struct {
	Task parser.Task `json:"task"`
}
type UpdateTaskRequest struct {
	ID     int `json:"id"`
	Result int `json:"result"`
}
type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	ID int `json:"id"`
}

type GetExpressionsRequest struct {
	Expressions []Expression `json:"expressions"`
}

var StorageTask []parser.Task
var StorageExpression []Expression

func PrintT() {
	for _, Task := range StorageTask {
		fmt.Println(Task)
	}
	fmt.Println()
}

func CalculateHandler(w http.ResponseWriter, r *http.Request) {
	request := new(CalculateRequest)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println("error:", err)
		return
	}
	response := CalculateResponse{ID: len(StorageExpression) + 1}
	StorageExpression = append(StorageExpression, Expression{
		ID:     response.ID,
		Status: "waiting",
		Result: "",
	})
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	taskMutex.Lock()
	postFormat, _ := parser.InPostcrementExpr(request.Expression)
	tasks := parser.DoTasks(postFormat, response.ID)
	StorageTask = append(StorageTask, tasks...)
	for _, task := range StorageTask {
		parser.PrintTask(task)
	}
	taskMutex.Unlock()
}

func GetExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	importedExpr := StorageExpression
	response := GetExpressionsResponse{
		Expressions: importedExpr,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func GetExpressionByIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idFromPath := vars["id"]

	id, err := strconv.Atoi(idFromPath)
	if err != nil {
		http.Error(w, "Invalid ID: must be an integer", http.StatusUnprocessableEntity)
		fmt.Printf("Invalid ID format: %s, error: %v\n", idFromPath, err)
		return
	}

	if id <= 0 || id-1 >= len(StorageExpression) {
		http.Error(w, "Expression not found", http.StatusNotFound)
		fmt.Printf("Expression with ID %d not found\n", id)
		return
	}
	expression := StorageExpression[id-1]

	response := GetExpressionByIdResponse{
		Expression: expression,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	taskMutex.Lock() // Блокируем доступ к StorageTask
	defer taskMutex.Unlock()

	if len(StorageTask) == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "No task available")
		return
	}

	var importedTask *parser.Task
	getTaskBool := false

	// Ищем первую задачу без зависимостей и сразу удаляем её
	for i := 0; i < len(StorageTask); i++ {
		task := &StorageTask[i]

		if len(task.Arg1) > 0 && task.Arg1[0] == 't' || len(task.Arg2) > 0 && task.Arg2[0] == 't' || task.Status == "waiting" {
			continue
		} else {
			importedTask = task
			importedTask.Status = "waiting"
			getTaskBool = true
			break
		}
	}

	if !getTaskBool {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "No task available")
		return
	}

	// Удаляем задачу из StorageTask
	//StorageTask = append(StorageTask[:taskIndex], StorageTask[taskIndex+1:]...)

	response := GetTaskResponse{Task: *importedTask}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

func UpdateTaskHandler(w http.ResponseWriter, r *http.Request) {
	request := new(UpdateTaskRequest)
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taskMutex.Lock() // Блокируем доступ к StorageTask
	defer taskMutex.Unlock()

	taskId := request.ID
	taskIdStr := strconv.Itoa(taskId)
	found := false

	if len(StorageTask) == 1 {
		for i := 0; i < len(StorageExpression); i++ {
			expression := &StorageExpression[i]
			if StorageTask[0].ExpressionId == expression.ID {
				temp := strconv.Itoa(request.Result)
				expression.Status = "done"
				expression.Result = temp
			}
		}
		StorageTask = []parser.Task{}
	}

	for i := 0; i < len(StorageTask); i++ {
		task := &StorageTask[i]
		if len(task.Arg1) > 0 && string(task.Arg1[0]) == "t" {
			arg1 := task.Arg1[1:]
			if arg1 == taskIdStr {
				task.Arg1 = strconv.Itoa(request.Result)
				found = true
				break
			}
		}
		if len(task.Arg2) > 0 && string(task.Arg2[0]) == "t" {
			arg2 := task.Arg2[1:]
			if arg2 == taskIdStr {
				task.Arg2 = strconv.Itoa(request.Result)
				found = true
				break
			}
		}
	}
	if found {
		for i := 0; i < len(StorageTask); i++ {
			if StorageTask[i].ID == taskId {
				StorageTask = append(StorageTask[:i], StorageTask[i+1:]...)
			}
		}
	}
	PrintT()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Task updated successfully")
}
