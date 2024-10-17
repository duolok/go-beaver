package main

import (
    "os"
	"log"
    "fmt"
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
        fmt.Printf("Task Name: %s\n", task.Name)
        fmt.Printf("File Path: %s\n", task.File)
        fmt.Printf("Schedule: Every %d %s\n", task.Schedule.Every, task.Schedule.Unit)
    }
}

