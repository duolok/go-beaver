package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"


	"github.com/fsnotify/fsnotify"
	"github.com/duolok/go-beaver/config"
	"github.com/duolok/go-beaver/taskrunner"
	yaml "gopkg.in/yaml.v3"
)

var (
	configFile = "config.yml"
)

func main() {
	config, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := setupGracefulShutdown()
	defer cancel()

    taskCancelFuncs := scheduleAllTasks(ctx, config)
    go watchConfigChanges(ctx, taskCancelFuncs)

	<-ctx.Done()
}

func scheduleAllTasks(ctx context.Context, config config.TaskConfig) []context.CancelFunc {
	var taskCancelFuncs []context.CancelFunc
	for _, task := range config.Tasks {
		duration, err := getDuration(task.Schedule.Every, task.Schedule.Unit)
		if err != nil {
			log.Fatalf("Invalid schedule unit for tasks %s: %v", task.Name, err)
		}
		taskCtx, cancel := context.WithCancel(ctx)
		go runScheduledTask(taskCtx, task, duration)
		taskCancelFuncs = append(taskCancelFuncs, cancel)
	}

	return taskCancelFuncs
}

func runScheduledTask(ctx context.Context, task config.Task, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
        select {
        case <- ctx.Done():
            fmt.Printf("Stopping task: %s\n", task.Name)
            return
        case <- ticker.C:
            fmt.Printf("Running task: %s\n", task.Name)
            taskrunner.RunTask(task.Type, task.File, interval)
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

				var newConfig config.TaskConfig
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
