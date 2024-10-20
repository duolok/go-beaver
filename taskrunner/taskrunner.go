package taskrunner

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
	"log"
    "time"
)


func RunTask(fileType string, filePath string, timeout time.Duration) {
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
