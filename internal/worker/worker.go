package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vizurth/concurency_calc/internal/server/handler"
	"log"
	"net/http"
	"strconv"
	"time"
)

func compute(arg1, arg2, operation string) (string, error) {
	a, err := strconv.Atoi(arg1)
	if err != nil {
		return "", err
	}
	b, err := strconv.Atoi(arg2)
	if err != nil {
		return "", err
	}

	switch operation {
	case "+":
		return strconv.Itoa(a + b), nil
	case "-":
		return strconv.Itoa(a - b), nil
	case "*":
		return strconv.Itoa(a * b), nil
	case "/":
		if b == 0 {
			return "", fmt.Errorf("division by zero")
		}
		return strconv.Itoa(a / b), nil
	default:
		return "", fmt.Errorf("unknown operation: %s", operation)
	}
}

func Worker(id int, host string, port string, client *http.Client) {
	if client == nil {
		log.Fatal("Worker: client is nil")
	}
	if host == "" || port == "" {
		log.Fatal("Worker: host or port is empty")
	}

	baseURL := fmt.Sprintf("http://%s:%s", host, port)

	log.Printf("Worker %d started with baseURL: %s", id, baseURL)

	for {
		resp, err := client.Get(baseURL + "/internal/task")

		if err != nil {
			log.Printf("Worker %d: Error getting task: %v", id, err)
			time.Sleep(1 * time.Second)
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			log.Printf("Worker %d: No tasks available", id)
			time.Sleep(1 * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Worker %d: Unexpected status: %d", id, resp.StatusCode)
			time.Sleep(1 * time.Second)
			continue
		}
		// Декодируем задачу
		var data handler.GetTaskResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			log.Printf("Worker %d: Error decoding task: %v", id, err)
			time.Sleep(1 * time.Second)
			continue
		}
		//fmt.Printf("Worker %d: Got task: %v\n", id, data)

		fmt.Printf("Worker %d: Received task %d: %s %s %s\n", id, data.Task.ID, data.Task.Arg1, data.Task.Operation, data.Task.Arg2)

		// Выполняем задачу
		duration, err := strconv.Atoi(data.Task.OperationTime)
		if err != nil {
			log.Printf("Worker %d: Invalid operation time for task %d: %v", id, data.Task.ID, err)
		}
		time.Sleep(time.Duration(duration) * time.Millisecond)

		result, err := compute(data.Task.Arg1, data.Task.Arg2, data.Task.Operation)

		if err != nil {
			log.Printf("Worker %d: Error computing task %d: %v", id, data.Task.ID, err)

			continue
		}
		resultInt, err := strconv.Atoi(result)
		if err != nil {
			log.Printf("Worker %d: Invalid result for task %d: %v", id, data.Task.ID, err)
		}
		// Отправляем результат
		payload := handler.UpdateTaskRequest{
			ID:     data.Task.ID,
			Result: resultInt, // Оставляем как string, так как UpdateTaskRequest ожидает string
		}
		jsonData, err := json.Marshal(&payload)
		if err != nil {
			log.Printf("Worker %d: Error marshaling payload: %v", id, err)
			continue
		}

		req, err := http.NewRequest("PUT", baseURL+"/internal/task", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Worker %d: Error creating request: %v", id, err)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = client.Do(req)
		if err != nil {
			log.Printf("Worker %d: Error sending result: %v", id, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Printf("Worker %d: Task %d completed with result %s", id, data.Task.ID, result)
		} else {
			log.Printf("Worker %d: Failed to update task %d, status: %d", id, data.Task.ID, resp.StatusCode)
		}
	}
}
