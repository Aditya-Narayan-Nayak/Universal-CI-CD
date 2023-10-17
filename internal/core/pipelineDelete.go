package core

import (
	"net/http"
	"net/url"
	"os"
)

type DeleteClient struct {
	workspacePath string
	logPath       string
	name          string
}

func (this *DeleteClient) NewClient(w http.ResponseWriter, r *http.Request) error {
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}

	// Get the value of the "name" parameter from the parsed query parameters
	pipeline := queryParams.Get("name")

	this.name = pipeline
	this.workspacePath = GenWorkspacePath(pipeline)
	this.logPath = GenLogFilePath(pipeline)

	return nil
}

func (this *DeleteClient) DeletePipeline() (int, error) {
	if err := os.RemoveAll(this.workspacePath); err != nil {
		return http.StatusInternalServerError, err
	}
	if err := os.RemoveAll(this.logPath); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
