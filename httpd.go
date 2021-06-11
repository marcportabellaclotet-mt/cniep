package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
)

type errorPageData struct {
	Code    string               `json:"code"`
	Title   string               `json:"title"`
	Details errorPageDataDetails `json:"details,omitempty"`
}

type errorPageDataDetails struct {
	OriginalURI               string            `json:"originalURI"`
	ContentType               string            `json:"contentType"`
	Namespace                 string            `json:"namespace"`
	IngressName               string            `json:"ingressName"`
	ServiceName               string            `json:"serviceName"`
	DeploymentName            string            `json:"deployName"`
	DesiredDeploymentReplicas int32             `json:"desiredDeploymentReplicas"`
	CurrentDeploymentReplicas int32             `json:"currentDeploymentReplicas"`
	ServicePort               string            `json:"servicePort"`
	RequestID                 string            `json:"requestId"`
	CustomErrorTemplate       string            `json:"customErrorTemplate"`
	CustomJsonResponse        map[string]string `json:"customJsonResponse"`
	CustomFields              map[string]string `json:"customFields"`
}

const (
	FormatHeader = "X-Format"
	CodeHeader   = "X-Code"
	ContentType  = "Content-Type"
	OriginalURI  = "X-Original-URI"
	Namespace    = "X-Namespace"
	IngressName  = "X-Ingress-Name"
	ServiceName  = "X-Service-Name"
	ServicePort  = "X-Service-Port"
	RequestID    = "X-Request-ID"
	JSON         = "application/json"
	HTML         = "text/html"
)

func webserver() {
	logrus.Println("Starting cniep httpd server (custom nginx ingress error pages)")
	http.HandleFunc("/", cniep)
	http.HandleFunc("/healthz", healthz)
	logrus.Fatal(http.ListenAndServe(":8082", nil))
}

func cniep(w http.ResponseWriter, r *http.Request) {
	requestInfo := newErrorPageData(r)
	originalURI := requestInfo.Details.OriginalURI
	if strings.HasSuffix(originalURI, "favicon.ico") {
		file := fmt.Sprintf("%v/favicon.ico", "/static/")
		if _, err := os.Stat(file); os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, file)
	}
	if strings.Contains(originalURI, "cniep-statics/") {
		filePath := fmt.Sprintf("/static/%v", strings.Split(originalURI, "cniep-statics/")[1])
		if fileExists(filePath) {
			file, err := os.Open(filePath)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			defer file.Close()
			_, filename := path.Split(filePath)
			http.ServeContent(w, r, filename, time.Time{}, file)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}
	code := requestInfo.Code
	codeInt, _ := strconv.Atoi(code)
	if codeInt == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	if requestInfo.Details.CustomJsonResponse[code] != "" {
		w.Header().Set(ContentType, JSON)
		body := requestInfo.Details.CustomJsonResponse[code]
		w.WriteHeader(codeInt)
		w.Write([]byte(body))
		return
	}
	if requestInfo.Details.CustomJsonResponse["global"] != "" {
		w.Header().Set(ContentType, JSON)
		body := requestInfo.Details.CustomJsonResponse["global"]
		w.WriteHeader(codeInt)
		w.Write([]byte(body))
		return
	}
	if strings.Contains(requestInfo.Details.OriginalURI, "cniep-svc-info") || strings.Contains(r.RequestURI, "cniep-svc-info") {
		w.Header().Set(ContentType, JSON)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Cniep-Response", "true")
		requestJson := unmarshallResource(&requestInfo)
		w.Write(requestJson)
		return
	}
	customErrorTemplate := requestInfo.Details.CustomErrorTemplate
	htmlFile, err := templateFileCheck(customErrorTemplate, code, "html")
	cssFile, err := templateFileCheck(customErrorTemplate, code, "css")

	f, err := os.Open(htmlFile)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	c, err := os.Open(cssFile)
	if err != nil {
		logrus.Error(err.Error())
		return
	}
	defer f.Close()
	defer c.Close()
	w.Header().Set(ContentType, HTML)
	tmpl := template.Must(template.ParseFiles(f.Name(), c.Name()))
	tmpl.Execute(w, requestInfo)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func newErrorPageData(req *http.Request) (errorPageData errorPageData) {
	code := req.Header.Get(CodeHeader)
	errorPageData.Code = code
	statusCodeNumber, _ := strconv.Atoi(code)
	errorPageData.Title = http.StatusText(statusCodeNumber)
	serviceName := req.Header.Get(ServiceName)
	namespace := req.Header.Get(Namespace)
	errorPageData.Details.ServiceName = serviceName
	errorPageData.Details.ServicePort = namespace
	errorPageData.Details.RequestID = req.Header.Get(RequestID)
	errorPageData.Details.IngressName = req.Header.Get(IngressName)
	errorPageData.Details.Namespace = req.Header.Get(Namespace)
	errorPageData.Details.OriginalURI = req.Header.Get(OriginalURI)
	errorPageData.Details.ContentType = req.Header.Get(ContentType)
	mapIdentifier := fmt.Sprintf("%v.%v", serviceName, namespace)
	errorPageData.Details.CustomErrorTemplate = serviceDetailsMap[mapIdentifier].customErrorTemplate
	errorPageData.Details.DeploymentName = serviceDetailsMap[mapIdentifier].deployName
	errorPageData.Details.CustomJsonResponse = serviceDetailsMap[mapIdentifier].customJsonResponse
	errorPageData.Details.CustomFields = serviceDetailsMap[mapIdentifier].customFields
	errorPageData.Details.DesiredDeploymentReplicas = serviceDetailsMap[mapIdentifier].desiredReplicas
	errorPageData.Details.CurrentDeploymentReplicas = serviceDetailsMap[mapIdentifier].currentReplicas
	return errorPageData
}
