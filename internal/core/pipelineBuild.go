package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"time"

	"gitlab.com/dipankardas011/universalci-cd/internal/utils"
)

type BuildClient struct {
	pipelineFile  *utils.UniversalFormat
	name          string
	logPath       string
	workspacePath string
	wHttp         utils.ResponseWriterFlusher
}

func (this *BuildClient) NewClient(w http.ResponseWriter, r *http.Request) error {
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}

	// Get the value of the "name" parameter from the parsed query parameters
	pipeline := queryParams.Get("name")
	this.logPath = GenLogFilePath(pipeline)
	this.workspacePath = GenWorkspacePath(pipeline)
	this.name = pipeline

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("Streaming not supported!")
	}

	this.wHttp = utils.ResponseWriterFlusher{ResponseWriter: w, Flusher: flusher}

	raw, err := os.ReadFile(this.workspacePath + "/" + utils.PIPELINE_FILE_NAME)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(raw, &this.pipelineFile); err != nil {
		return err
	}
	return nil
}

func (this *BuildClient) PipelineHandler() error {

	rawBuildNo, err := getBuildNumber(this.logPath)
	if err != nil {
		return err
	}
	buildNo := strconv.Itoa(rawBuildNo)

	if err := createWorkspaceDir(this.logPath + "/" + buildNo); err != nil {
		return err
	}

	// Create files to store the output
	outputStderr, err := os.Create(this.logPath + "/" + buildNo + "/" + utils.LOG_FILE_ERR)
	if err != nil {
		return err
	}
	defer outputStderr.Close()

	outputStdout, err := os.Create(this.logPath + "/" + buildNo + "/" + utils.LOG_FILE_OUT)
	if err != nil {
		return err
	}
	defer outputStdout.Close()

	metadataStdout, err := os.Create(this.logPath + "/" + buildNo + "/" + utils.METADATA_BUILD)
	if err != nil {
		return err
	}
	defer metadataStdout.Close()

	this.wHttp.ResponseWriter.Header().Set("Buildno", strconv.Itoa(rawBuildNo))

	this.wHttp.Write([]byte(fmt.Sprintf("\t Pipeline -> %s\n", this.name)))
	this.wHttp.Write([]byte(fmt.Sprintf("\t BuildNumber -> %d\n------\n", rawBuildNo)))

	var pipelineError error

	stages := this.pipelineFile.Stages

	start := time.Now()
stage:
	for _, stage := range stages {

		this.wHttp.Write([]byte(fmt.Sprintf("\tExecuting Stage [[[ %s ]]]\n\n", stage.Name)))
		outputStdout.Write([]byte(fmt.Sprintf("\tExecuting Stage [[[ %s ]]]\n\n", stage.Name)))
		jobs := stage.Jobs

		for _, job := range jobs {

			rawScript := job.Script
			dir, err := os.MkdirTemp("", "universal-runnner*")
			defer os.RemoveAll(dir)

			if err != nil {
				pipelineError = err
				break stage
			}
			scriptPath := dir + "/job.sh"

			this.wHttp.Write([]byte(fmt.Sprintln(scriptPath)))
			outputStdout.Write([]byte(fmt.Sprintln(scriptPath)))

			if err := os.WriteFile(scriptPath, []byte(rawScript), 0510); err != nil {
				this.wHttp.Write([]byte(fmt.Sprintf("[ FAILED ]: %v", err)))
				outputStderr.Write([]byte(fmt.Sprintf("[ FAILED ]: %v", err)))
				pipelineError = err
				break stage
			}
			if err := runCommand(this.workspacePath, scriptPath, this.wHttp, outputStdout, outputStderr); err != nil {
				this.wHttp.Write([]byte(fmt.Sprintf("[ FAILED ]: %v", err)))
				outputStderr.Write([]byte(fmt.Sprintf("[ FAILED ]: %v", err)))
				pipelineError = err
				break stage
			}
		}
	}

	pipelineMetadata := utils.MetadataBuild{}
	pipelineMetadata.TimeTaken = time.Since(start).String()
	pipelineMetadata.BuildNo = rawBuildNo

	if pipelineError != nil {
		pipelineMetadata.Status = "FAILED"
	} else {
		pipelineMetadata.Status = "SUCCESS"
		outputStdout.Write([]byte("[ SUCCESS ]"))
		this.wHttp.Write([]byte("[ SUCCESS ]"))
	}
	raw, err := json.Marshal(pipelineMetadata)
	if err != nil {
		return err
	}
	if _, err := metadataStdout.Write(raw); err != nil {
		return err
	}
	return pipelineError
}

func GenWorkspacePath(pipeline string) string {
	return utils.WORKSPACE_DIR + pipeline
}

func GenLogFilePath(pipeline string) string {
	return utils.LOG_DIR + pipeline
}

func createWorkspaceDir(dir string) error {

	// Check if the directory already exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory doesn't exist, so create it
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}
		fmt.Println("Directory created:", dir)
	} else if err != nil {
		fmt.Println("Error checking directory existence:", err)
		return err
	} else {
		fmt.Println("Directory already exists:", dir)
	}
	return nil
}

func getBuildNumber(dir string) (int, error) {
	// Read the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println("Error:", err)
		return 0, err
	}

	// Create a slice to store folder names as integers
	folderNames := []int{}

	// Iterate through the directory entries
	for _, entry := range entries {
		// Check if it's a directory
		if entry.IsDir() {
			// Try to convert the folder name to an integer
			folderName, err := strconv.Atoi(entry.Name())
			if err == nil {
				folderNames = append(folderNames, folderName)
			}
		}
	}
	BuildNumberPresent := 0
	for _, build := range folderNames {
		if build > BuildNumberPresent {
			BuildNumberPresent = build
		}
	}
	return BuildNumberPresent + 1, nil
}

func runCommand(workingDir string, path string, wHTTP utils.ResponseWriterFlusher, outputStdout, outputStderr *os.File) error {
	cmd := exec.Command("bash", path)

	cmd.Dir = workingDir

	// Create pipes for the command's standard output and standard error
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	// Create goroutines to continuously read and print the output
	go func() {
		// defer outputStdout.Close() // Close the file when the goroutine exits
		// io.Copy(io.MultiWriter(os.Stdout, wHTTP, outputStdout), stdoutPipe)
		io.Copy(io.MultiWriter(wHTTP, outputStdout), stdoutPipe)
		outputStdout.Sync() // Flush the data to the file
	}()

	go func() {
		// defer outputStderr.Close() // Close the file when the goroutine exits
		io.Copy(io.MultiWriter(wHTTP, outputStderr), stderrPipe)
		outputStderr.Sync() // Flush the data to the file
	}()

	// Wait for the command to complete
	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}
