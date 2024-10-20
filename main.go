package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	yaml "gopkg.in/yaml.v3"
)

type Task struct {
	File     string `yaml:"file"`
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Schedule struct {
		Every int    `yaml:"every"`
		Unit  string `yaml:"unit"`
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

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	ctx, cancel := setupGracefulShutdown()
	defer cancel()

	var config TaskConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
    
    taskCancelFuncs := scheduleAllTasks(ctx, config)

    go watchConfigChanges(ctx, taskCancelFuncs)


	<-ctx.Done()
}

func scheduleAllTasks(ctx context.Context, config TaskConfig) []context.CancelFunc {
	var taskCancelFuncs []context.CancelFunc
	for _, task := range config.Tasks {
		duration, err := getDuration(task.Schedule.Every, task.Schedule.Unit)
		if err != nil {
			log.Fatalf("Invalid schedule unit for tasks %s: %v", task.Name, err)
		}
		taskCtx, cancel := context.WithCancel(ctx)
		go scheduleTask(taskCtx, task, duration)
		taskCancelFuncs = append(taskCancelFuncs, cancel)
	}

	return taskCancelFuncs
}

func scheduleTask(ctx context.Context, task Task, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
        select {
        case <- ctx.Done():
            fmt.Printf("Stopping task: %s\n", task.Name)
            return
        case <- ticker.C:
            fmt.Printf("Running task: %s\n", task.Name)
            runTask(task.Type, task.File, interval)
        }
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

	log.Printf("Task %s output: \n%s", filePath, string(output))
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

func watchConfigChanges(ctx context.Context, oldTaskCancelFuncs []context.CancelFunc) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Error creating watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(configFile)
	if err != nil {
		log.Fatalf("Error watching config file: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("config file modified, reloading tasks...")

				for _, cancel := range oldTaskCancelFuncs {
					cancel()
				}

				data, err := os.ReadFile(configFile)
				if err != nil {
					log.Printf("Error reading config file: %v", err)
					continue
				}

				var newConfig TaskConfig
				err = yaml.Unmarshal(data, &newConfig)
				if err != nil {
					log.Printf("Error parsing config file: %v", err)
					continue
				}

				oldTaskCancelFuncs = scheduleAllTasks(ctx, newConfig)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)

		}
	}
}

func setupGracefulShutdown() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %s. Shutting down...\n", sig)
		cancel()
	}()

	return ctx, cancel
}
