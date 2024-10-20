package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v3"
)

type Task struct {
    File    string `yaml:"file"`
    Name    string `yaml:"name"`
    Type    string `yaml:"type"`
    Schedule struct {
        Every int   `yaml:"every"`
        Unit  string   `yaml:"unit"`
    } `yaml:"schedule"`
}

type TaskConfig struct {
    Tasks []Task `yaml:"tasks"`
}

var (
    configFile = "config.yml"
)

func main() {
    data, err := os.ReadFile(configFile)
    if err != nil {
        log.Fatalf("An error has occured.")
    }

    var config TaskConfig
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        log.Fatalf("error: %v", err)
    }

    for _, task := range config.Tasks {
        duration, err := getDuration(task.Schedule.Every, task.Schedule.Unit)
        if err != nil {
            log.Fatalf("Invalid schedule unit for task %s: %v", task.Name, err)
        }

        go scheduleTask(task, duration)
    }

    select {}
}

func scheduleTask(task Task, interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    for {
        <-ticker.C
        fmt.Printf("Running task: %s\n", task.Name)
        runTask(task.Type, task.File, interval)
    }
}

func getDuration(every int, unit string) (time.Duration, error) {
    switch unit {
    case "seconds":
        return time.Duration(every) * time.Second, nil
    case "minutes":
        return time.Duration(every) * time.Minute, nil
    case "hours":
        return time.Duration(every) * time.Hour, nil
    default:
        return 0, fmt.Errorf("Unsupported time unit: %s", unit)
    }
}

func runTask(fileType string, filePath string, timeout time.Duration) {
    taskFileType, err := handleScriptType(fileType) 
    if err != nil {
        log.Printf("Unknown error type")
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()

    cmd := exec.CommandContext(ctx, taskFileType, filePath)

    output, err := cmd.CombinedOutput()
    if ctx.Err() == context.DeadlineExceeded {
        log.Printf("Task %s timed out after %v\n", filePath, timeout)
        return
    }

    if err != nil {
        log.Printf("Error running task %s: %v", filePath, err)
        log.Printf("Output: \n%s", string(output))
        return
    }

    log.Printf("Task %s output: \n%s",filePath, string(output))
    fmt.Printf("Task %s completed successfully\n", filePath)
}

func handleScriptType(fileType string) (string, error) {
    fileType = strings.TrimSpace(strings.ToLower(fileType))

    switch fileType {
    case "sh":
        return "bash", nil
    case "python":
        return "python3", nil
    case "bin":
        return "./", nil
    default:
        return "0", fmt.Errorf("unknown filetype")
    }
}
