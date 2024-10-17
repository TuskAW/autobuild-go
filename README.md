# Autobuild-Go

Autobuild-Go is a command-line tool designed to automate the testing and building of Go projects. It simplifies the process of ensuring that `Go` is installed, walking through the project directories, identifying Go projects, and building them.

## Features

- **Project Discovery**: Automatically walks through a directory to find Go projects by locating `main.go` and `go.mod` files.
- **Go Installation Check**: Ensures that Go is installed on the system before running the build process.
- **Automated Build**: Uses a builder to test and build Go projects, managing output and error handling efficiently.

## Installation

### From releases

To install autobuild-go without need of building it from scratch you can use provided releases here: https://github.com/mateuszmierzwinski/autobuild-go/releases

### From source

To install Autobuild-Go, you can clone the repository and build the tool using Go.

```bash
git clone https://github.com/mateuszmierzwinski/autobuild-go.git
cd autobuild-go
go build -o autobuild-go ./cmd/autobuild-go/
```

Make sure Go is installed on your system. You can install Go by following the official [Go installation guide](https://golang.org/doc/install).

## Usage

You can run the tool with the following command:

```bash
./autobuild-go [path]
```

- `path`: (Optional) The root directory you want the tool to scan for Go projects.
- If no path is provided, the tool will default to the **current directory** and scan it for Go projects.

### Example

```bash
./autobuild-go /path/to/projects
```

If no path is specified, it will search for Go projects in the current directory:

```bash
./autobuild-go
```

## How it Works

1. **ProjectWalker**: The tool starts by walking through the provided path (or the current directory if none is provided), looking for Go projects by identifying `main.go` and pairing it with the closest `go.mod` file.
2. **GoBuilder**: Once projects are identified, the GoBuilder takes over, ensuring Go is installed and running `go build` on each project.
3. **Parallel Processing**: The tool is designed to handle multiple projects simultaneously using Go channels and goroutines.

## Project Structure

- `main.go`: Entry point for the tool. Ensures Go installation and initiates the project walking and building process.
- `projectwalker.go`: Contains logic for walking through directories and identifying Go projects.
- `gobuilder.go`: Responsible for testing and building the identified Go projects.
- `processor.go`: Manages the processing pipeline, initiating the project walker and builder.
- `golanginstaller.go`: Ensures Go is installed before proceeding with any build operations.
- `version.go`: Contains versioning information for the tool.

## License

This project is licensed under the BSD 2-Clause License - see the [LICENSE](./LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request if you have any improvements or fixes.
