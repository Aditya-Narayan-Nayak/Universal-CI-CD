package utils

import (
	"net/http"
)

type Job struct {
	Script string `json:"script"`
}

type Stage struct {
	Name string `json:"name"`
	Jobs []Job  `json:"jobs"`
}

type UniversalFormat struct {
	PipelineName string   `json:"name"`
	Agents       []string `json:"agents"`
	Stages       []Stage  `json:"stages"`
}

type ResponseWriterFlusher struct {
	http.ResponseWriter
	http.Flusher
}

func (rw ResponseWriterFlusher) Write(p []byte) (n int, err error) {
	n, err = rw.ResponseWriter.Write(p)
	rw.Flusher.Flush()
	return n, err
}

type MetadataBuild struct {
	BuildNo   int
	Status    string
	TimeTaken string
}
