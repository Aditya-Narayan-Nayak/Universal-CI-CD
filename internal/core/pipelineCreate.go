package core

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"gitlab.com/dipankardas011/universalci-cd/internal/utils"
)

type CreateClient struct {
	pipelineFile  *utils.UniversalFormat
	workspacePath string
	logPath       string
	name          string
}

func (this *CreateClient) NewClient(w http.ResponseWriter, r *http.Request) error {
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return err
	}

	// Get the value of the "name" parameter from the parsed query parameters
	pipeline := queryParams.Get("name")

	this.name = pipeline
	this.workspacePath = GenWorkspacePath(pipeline)
	this.logPath = GenLogFilePath(pipeline)

	this.pipelineFile = &utils.UniversalFormat{}

	if err := json.NewDecoder(r.Body).Decode(&this.pipelineFile); err != nil {
		return err
	}
	return nil
}

func (this *CreateClient) CreatePipeline() (int, error) {
	if err := createWorkspaceDir(this.workspacePath); err != nil {
		return http.StatusInternalServerError, err
	}
	if err := createWorkspaceDir(this.logPath); err != nil {
		return http.StatusInternalServerError, err
	}

	raw, err := json.Marshal(this.pipelineFile)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if err := os.WriteFile(this.workspacePath+"/"+utils.PIPELINE_FILE_NAME, raw, 0640); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
