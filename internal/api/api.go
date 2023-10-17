package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"gitlab.com/dipankardas011/universalci-cd/internal/core"
	"gitlab.com/dipankardas011/universalci-cd/internal/utils"
)

type apiFunc func(http.ResponseWriter, *http.Request) (int, error)

func MakeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Printf("[%s] %s ⚡", r.Method, r.URL.Path)
		start := time.Now()

		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Buildno") // Add this line

		statCode, err := f(w, r)
		if err != nil {
			log.Println(err)
			w.WriteHeader(statCode)
			w.Write([]byte(err.Error()))
		}

		log.Printf("[%s] %s {%d} %v", r.Method, r.URL.Path, statCode, time.Since(start))
	}
}

func Docs(w http.ResponseWriter, r *http.Request) (int, error) {

	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed, fmt.Errorf("GET method is allowed")
	}

	type Param struct {
		Param     string
		ValueType string
		Desc      string
	}

	type Doc struct {
		Method   string
		Endpoint string
		Params   []Param
		ReqData  string
	}
	docs := []Doc{
		Doc{
			Method:   http.MethodPost,
			Endpoint: "/create",
			ReqData:  "json data of the pipeline",
			Params: []Param{
				Param{
					Param:     "name",
					ValueType: "string",
					Desc:      "name of the pipeline",
				},
			},
		},

		Doc{
			Method:   http.MethodGet,
			Endpoint: "/list",
		},

		Doc{
			Method:   http.MethodGet,
			Endpoint: "/build",
		},

		Doc{
			Method:   http.MethodGet,
			Endpoint: "/get",
			Params: []Param{
				Param{
					Param:     "name",
					ValueType: "string",
					Desc:      "name of the pipeline",
				},
				Param{
					Param:     "build",
					ValueType: "number",
					Desc:      "build number",
				},
				Param{
					Param:     "detail",
					ValueType: "boolean",
					Desc:      "whether to return only boolean or detail [CURRENTLY] only metadata for the builds are kept not forthe pipeline",
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(docs); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil

}

func BuildNow(w http.ResponseWriter, r *http.Request) (int, error) {

	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed, fmt.Errorf("GET method is allowed")
	}

	abcd := &core.BuildClient{}
	if err := abcd.NewClient(w, r); err != nil {
		return http.StatusInternalServerError, err
	}

	if err := abcd.PipelineHandler(); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func DeletePipeline(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method != http.MethodPut {
		return http.StatusMethodNotAllowed, fmt.Errorf("PUT method is allowed")
	}
	abcd := &core.DeleteClient{}
	if err := abcd.NewClient(w, r); err != nil {
		return http.StatusInternalServerError, err
	}
	return abcd.DeletePipeline()
}

func CreatePipeline(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, fmt.Errorf("POST method is allowed")
	}

	abcd := &core.CreateClient{}
	if err := abcd.NewClient(w, r); err != nil {
		return http.StatusInternalServerError, err
	}
	return abcd.CreatePipeline()
}

func GetAllPipelines(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed, fmt.Errorf("GET method is allowed")
	}

	read, err := os.ReadDir(utils.WORKSPACE_DIR)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type Workspaces struct {
		Folders []string
	}
	workspace := Workspaces{}

	for _, dirEntry := range read {
		if dirEntry.IsDir() {
			workspace.Folders = append(workspace.Folders, dirEntry.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(workspace); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func GetBuildDetails(w http.ResponseWriter, r *http.Request) (int, error) {

	if r.Method != http.MethodGet {
		return http.StatusMethodNotAllowed, fmt.Errorf("GET method is allowed")
	}
	queryParams, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return http.StatusNotFound, err
	}

	// Get the value of the "name" parameter from the parsed query parameters
	pipeline := queryParams.Get("name")
	buildNo := queryParams.Get("build")
	detail := queryParams.Get("detail")

	if len(buildNo) == 0 {

		read, err := os.ReadDir(utils.LOG_DIR + pipeline)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		type Workspaces struct {
			Builds []string
		}
		builds := Workspaces{}

		for _, dirEntry := range read {
			if dirEntry.IsDir() {
				builds.Builds = append(builds.Builds, dirEntry.Name())
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(builds); err != nil {
			return http.StatusInternalServerError, err
		}
	} else {

		buildFolder := core.GenLogFilePath(pipeline) + "/" + buildNo + "/"

		if len(detail) == 0 {

			contentOfMetadata, err := os.ReadFile(buildFolder + utils.METADATA_BUILD)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			metadata := utils.MetadataBuild{}

			if err := json.Unmarshal(contentOfMetadata, &metadata); err != nil {
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(metadata); err != nil {
				return http.StatusInternalServerError, err
			}

		} else if detail == "true" {
			contentOfStdOut, err := os.ReadFile(buildFolder + utils.LOG_FILE_OUT)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			contentOfStdErr, err := os.ReadFile(buildFolder + utils.LOG_FILE_ERR)
			if err != nil {
				return http.StatusInternalServerError, err
			}

			contentOfMetadata, err := os.ReadFile(buildFolder + utils.METADATA_BUILD)
			if err != nil {
				return http.StatusInternalServerError, err
			}
			type LogsBuild struct {
				Stdout   string
				Stderr   string
				Metadata utils.MetadataBuild
			}
			logs := LogsBuild{}

			logs.Stdout = string(contentOfStdOut)
			logs.Stderr = string(contentOfStdErr)

			if err := json.Unmarshal(contentOfMetadata, &logs.Metadata); err != nil {
				return http.StatusInternalServerError, err
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(logs); err != nil {
				return http.StatusInternalServerError, err
			}

		}
	}

	return http.StatusOK, nil
}
