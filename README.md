# Go-Beaver

Go Beaver is a highly flexible and modular task scheduler built with Go. It enables you to schedule, execute, and monitor tasks (e.g., shell scripts, Python scripts) based on time intervals specified in a configuration file.

## 🚀 Features
- **Task Scheduling**: Run tasks like shell scripts, Python scripts, or binaries at defined intervals (seconds, minutes, hours).
- **Live Config Reloading**: Modify your task schedule without restarting the application. Config changes are detected on the fly.
- **Graceful Shutdown**: All running tasks are gracefully terminated when the app is stopped.
- **Task Execution with Timeout**: Configure timeouts for tasks to prevent long-running processes from blocking the system.

## 📖 Table of Contents
- [Getting Started](#-getting-started)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [Usage](#-usage)
- [Contributing](#-contributing)
- [License](#-license)

## 🏁 Getting Started

Follow these instructions to set up and run Go Beaver on your local machine.

### Prerequisites
- [Go 1.19+](https://golang.org/dl/) installed on your machine.
- Basic understanding of YAML for configuration.

### Installation

Clone the repository:

```bash
git clone https://github.com/your-username/go-beaver.git
cd go-beaver
go mod tidy
go build -o go-beaver
./go-beaver
```

### Configuration
Go Beaver uses a YAML file (config.yml) to configure task schedules. Below is an example of a config.yml file:

```yaml
tasks:
  - name: "Backup Task"
    file: "/path/to/backup.sh"
    type: "sh"
    schedule:
      every: 30
      unit: "minutes"

  - name: "Python Task"
    file: "/path/to/script.py"
    type: "python"
    schedule:
      every: 2
      unit: "hours"
```
## Task Configuration Options
 - name: A unique name for the task.
 - file: The path to the file you want to execute.
 - type: The type of file (e.g., sh, python, bin).
 - schedule.every: The interval at which the task should run.
 - schedule.unit: Time unit (seconds, minutes, hours).

## Usage
Once the application is running, Go Beaver will automatically schedule tasks defined in your config.yml file. It will also monitor the configuration file for any changes and reload tasks dynamically.

#### Stopping the App

Use CTRL+C to stop the app. Go-Beaver will ensure all tasks are properly terminated before exiting.
