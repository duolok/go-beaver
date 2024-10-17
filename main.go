package main

import (
	"fmt"
	"log"
	"os"
    "os/exec"
	"time"

	yaml "gopkg.in/yaml.v3"
)

type Task struct {
    File    string `yaml:"file"`
    Name    string `yaml:"name"`
    Schedule struct {
        Every int   `yaml:"every"`
        Unit  string   `yaml:"unit"`
    } `yaml:"schedule"`
}

type TaskConfig struct {
    Tasks []Task `yaml:"tasks"`
}

func main() {
    data, err := os.ReadFile("config.yml")
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
        runTask(task.File)
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

func runTask(filePath string) {
    cmd := exec.Command("python3", filePath)
    if err := cmd.Start(); err != nil {
        log.Printf("Error starting task %s: %v", filePath, err)
        return
    }

    if err := cmd.Wait(); err != nil {
        log.Printf("Task %s failed: %v", filePath, err)
        return
    }

    fmt.Printf("Task %s completed successfully\n", filePath)
}
