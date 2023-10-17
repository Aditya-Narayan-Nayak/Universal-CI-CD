package ui

import (
	"html/template"
	"net/http"
)

func DeleteUI(w http.ResponseWriter, r *http.Request) (int, error) {
	t, err := template.ParseFiles("../template/indexDelete.html", "../template/common.html", "../template/scriptDelete.html", "../template/styles.html")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type Home struct {
		Operation string
		Content   template.HTML
	}

	abcd := Home{
		Operation: "Delete",
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, abcd); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func CreateUI(w http.ResponseWriter, r *http.Request) (int, error) {
	t, err := template.ParseFiles("../template/indexCreate.html", "../template/common.html", "../template/scriptCreate.html", "../template/styles.html")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type Home struct {
		Operation string
		Content   template.HTML
	}

	abcd := Home{
		Operation: "Create",
		Content: template.HTML(`
{
	"name": "Pipeline name",
	"agents": [
		"any"
	],
	"stages": [
		{
			"name": "git changelog",
			"jobs": [
				{
					"script": "#!/bin/bash\necho \"Git changelog\""
				}
			]
		},
		{
			"name": "Build Ksctl",
			"jobs": [
				{
					"script": "#!/bin/bash\necho \"Building\""
				}
			]
		}
	]
}
`),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, abcd); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func ListUI(w http.ResponseWriter, r *http.Request) (int, error) {
	t, err := template.ParseFiles("../template/indexList.html", "../template/common.html", "../template/scriptList.html", "../template/styles.html")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type Home struct {
		Operation string
		Content   template.HTML
	}

	abcd := Home{
		Operation: "List",
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, abcd); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func HomeUI(w http.ResponseWriter, r *http.Request) (int, error) {
	t, err := template.ParseFiles("../template/index.html", "../template/common.html", "../template/styles.html")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	type Home struct {
		Operation string
		Content   template.HTML
	}

	abcd := Home{
		Operation: "Home",
		Content: template.HTML(`
<h3>Purpose</h3>
<p>To Demonstrate a simple CI server. It can perform creation of pipeline, Build Now and check the previous build with status</p>
<p>future goals: make it more refined and as this has both ui and an api backend we can implement auth and rust based cli</p>
<hr>
<div>Author: Dipankar Das</div>
`),
	}

	w.Header().Set("Content-Type", "text/html")
	if err := t.Execute(w, abcd); err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}
