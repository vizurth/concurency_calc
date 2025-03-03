package parser

import (
	"container/list"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Task struct {
	ID            int    `json:"id"`
	Arg1          string `json:"arg1"`
	Arg2          string `json:"arg2"`
	Operation     string `json:"operation"`
	OperationTime string `json:"operation_time"`
	Status        string `json:"status"`
	ExpressionId  int    `json:"expression_id"`
}

type TimeConfig struct {
	TimeAddition    string
	TimeSubtraction string
	TimeMultiplier  string
	TimeDivision    string
}

func envTime() *TimeConfig {
	timeAdd := os.Getenv("TIME_ADDITION_MS")
	if timeAdd == "" {
		timeAdd = "1"
	}
	timeSub := os.Getenv("TIME_SUBTRACTION_MS")
	if timeSub == "" {
		timeSub = "1"
	}
	timeMul := os.Getenv("TIME_MULTIPLICATIONS_MS")
	if timeMul == "" {
		timeMul = "2"
	}
	timeDiv := os.Getenv("TIME_DIVISION_MS")
	if timeDiv == "" {
		timeDiv = "2"
	}
	return &TimeConfig{timeAdd, timeSub, timeMul, timeDiv}
}

func InPostcrementExpr(expression string) ([]string, error) {
	priority := map[string]int{
		"+": 1,
		"-": 1,
		"*": 2,
		"/": 2,
	}

	re := regexp.MustCompile(`\d+|[+\-*/()]`)
	expr := re.FindAllString(expression, -1)

	stack := list.New()
	queue := make([]string, 0)

	if !currBrackets(expression) {
		return []string{}, fmt.Errorf("expression '%s' contains uncurrect brackets format", expression)
	}

	for _, token := range expr {
		fmt.Println(token)
		if isNumber(token) {
			queue = append(queue, token)
		} else if isOperation(token) {
			for stack.Len() > 0 {
				top := stack.Back().Value.(string)
				if isOperation(top) && priority[top] >= priority[token] {
					queue = append(queue, top)
					stack.Remove(stack.Back())
				} else {
					break
				}
			}
			stack.PushBack(token)
		} else if token == "(" {
			stack.PushBack(token)
		} else if token == ")" {
			for stack.Len() > 0 && stack.Back().Value.(string) != "(" {
				queue = append(queue, stack.Remove(stack.Back()).(string))
			}
			if stack.Len() > 0 {
				stack.Remove(stack.Back())
			}
		}
	}

	for stack.Len() > 0 {
		queue = append(queue, stack.Remove(stack.Back()).(string))
	}

	return queue, nil
}

func DoTasks(queue []string, expressionId int) []Task {
	tasks := make([]Task, 0)
	times := envTime()
	timeMap := map[string]string{
		"+": times.TimeAddition,
		"-": times.TimeSubtraction,
		"*": times.TimeMultiplier,
		"/": times.TimeDivision,
	}
	queueNew := make([]string, 0)
	queueNew = append(queueNew, queue...)
	for len(queueNew) > 3 {
		if len(queueNew) == 1 && string(queueNew[0][0]) == "t" {
			break
		}
		index := 0
		for range queueNew {
			if len(queueNew) == 3 {
				tasks = append(tasks, Task{
					ID:            len(tasks) + 1,
					Arg1:          queueNew[0],
					Arg2:          queueNew[1],
					Operation:     queueNew[2],
					OperationTime: timeMap[queueNew[2]],
					Status:        "open",
					ExpressionId:  expressionId,
				})
				break
			}
			if isOperation(queueNew[index]) && index >= 2 {
				tasks = append(tasks, Task{
					ID:            len(tasks) + 1,
					Arg1:          queueNew[index-2],
					Arg2:          queueNew[index-1],
					Operation:     queueNew[index],
					OperationTime: timeMap[queueNew[index]],
					Status:        "open",
					ExpressionId:  expressionId,
				})
				rightPart := make([]string, 0)
				rightPart = append(rightPart, queueNew[index+1:]...)

				addElem := fmt.Sprintf("t%d", len(tasks))
				queueNew = append(queueNew[:index-2], addElem)
				queueNew = append(queueNew, rightPart...)
				index = 0
			}
			index++

		}
	}

	return tasks
}

func isNumber(token string) bool {
	_, err := strconv.ParseFloat(token, 64)
	return err == nil
}

func isOperation(op string) bool {
	return op == "+" || op == "-" || op == "*" || op == "/"
}

func currBrackets(expression string) bool {
	expr := strings.TrimSpace(expression)
	count := 0
	for i := 0; i < len(expr); i++ {
		if expr[i] == '(' {
			count++
		} else if expr[i] == ')' {
			count--
		}
	}
	return count == 0
}

func PrintTask(task Task) {
	fmt.Printf("ID: %d\n Arg1: %s\n Arg2: %s\n Opetation: %s\n", task.ID, task.Arg1, task.Arg2, task.Operation)
}
