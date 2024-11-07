# deto - Dev Tools

Includes all common tools for dev on command line

cmd/: Contains the main applications of the project. Each folder under cmd is usually a separate executable. The main.go file in each folder is the entry point for that particular app.

pkg/: Holds utility libraries or reusable packages across your project. Any functionality you want to share across your project, like custom loggers or helpers, goes here.

configs/: Stores configuration files, such as YAML or JSON files, for setting up the application. This is useful for environment settings or loading configuration at runtime.

scripts/: Contains various shell scripts or build scripts. These might automate tasks like testing, deployment, or database migrations.

# Cobra CLI Library
Adding new commands to the CLI is easy with Cobra. To add a new command, you can use the Cobra CLI generator. This will create a new command file in the cmd/ directory with a basic structure.
```bash
cobra-cli add [commandName] --config ./configs/cobra.yaml
```

# Building the CLI
To build the CLI, you can use the go build command. This will compile the application and create an executable file in the root directory of the project.
```bash
go build -o deto
```

