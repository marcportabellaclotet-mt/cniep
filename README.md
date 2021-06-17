<p align="center">
  <img src="https://raw.githubusercontent.com/marcportabellaclotet-mt/cniep/master/static/images/CNIEP-2.png"></img>
</p>
<hr>
<h3 align="center">CNIEP is a replacement for default nginx backend controller to display customized error pages</h3>
<hr>

## Table of Contents

- [Description](#description)

- [Installation](#installation)

- [Enable CNIEP In K8s Services](#enable-cniep-for-k8s-services)

- [Supported Annotations](#supported-annotations)

- [How to template](#how-to-template)

- [Supported template variables](#template-variables)

- [Sync templates from S3 buckets](#sync-templates-from-s3-buckets)

- [Build from source](#build-from-source)

---
## Description

**cniep** ( custom nginx ingress error pages ) is a web service coupled to nginx ingress controller designed to provide customized error pages.

**cniep** traps pages returning error responses and displays a customized error page.

To display custom error responses some [ingress annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#custom-http-errors) are required.

**cniep** gets k8s deployments information which can be used to render custom error response pages.

More info about about custom error codes can be found [here](https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/).

---
## Installation

1 - Navigate to **cniep** kubernetes folder and customize manifests to fit your setup.

2 - Deploy **cniep**

<details><summary>Deployment command</summary>

``` 
# kubectl apply -f kubernetes 
```
</details>

3 - Modify your nginx controller to use cniep as default backend.

4 - Assuming that cniep is installed in cniep k8s namespace, add default-backend-service arg in your ingress nginx controller deployment.

<details><summary>Ingress Deployment args</summary>

```
      ...
      containers:
      - args:
        - /nginx-ingress-controller
        - --default-backend-service=cniep/cniep
      ...  
```
</details>

---
## Enable cniep for your services
To enable cniep for your services, follow the next steps:

1 - Install cniep deployment in your k8s server (check [installation](#installation) section)

2 - Annotate your k8s services and ingresses.

<details><summary>Service annotations</summary>

```
apiVersion: v1
kind: Service
metadata:
  name: myservice
  annotations:
    cniep/deployment: myservice
    cniep/template: cniep
```
</details>

<details><summary>Ingress annotations</summary>

3 - Annotate your k8s ingress definition. i.e:
```
apiVersion: extensions/v1beta1
kind: Ingress
metadata:
  name: myservice
  annotations:
    kubernetes.io/ingress.class: nginx
    nginx.ingress.kubernetes.io/custom-http-errors: "503"
```
</details>

---
## Supported Annotations

| Annotation | Description |
| --- | --- |
| `cniep/template` | Defines the template used to render the error response pages|
| `cniep/deployment` | Defines the deployment backend for the annotated service. If this annotation is not set, it will search for service selectors|
| `customjsonresponse-xxx` | Renders a custom json response instead of a templated html page (status code specific)|
| `customjsonresponse-global` | Renders a custom json response instead of a templated html page (all status codes)|
| `cniep/s3-templatedir` |  Syncs templates stored in an AWS S3 bucket. i.e. cniep/s3-templatedir: s3://mys3bucket/cniep2|
| `cniep/customfield-xxxx` | Define custom variables to be used in error response pages. In the [template](how-to-template) section, there is more information about this|
| `cniep/forceresponsecode-xxx` | Forces the HTTP status code returned by the request |

---

## How to template
**cniep** uses golang [text/template](https://golang.org/pkg/text/template/) package to render error pages.

<details><summary>Template folder structure</summary>

```
mytemplate (template folder)
 - index.html (html file)
 - style.css (css style file)
```
</details>
<details><summary>Template folder structure for specific status codes</summary>

```
mytemplate (template folder)
 - index-503.html (html file)
 - style-503.css (css style file)
 - index.html (html file)
 - style.css (css style file)
```
</details>
<details><summary>index.html</summary>

```
<!DOCTYPE  html>
<html>
<head>
<meta  http-equiv="Content-Type"  content="text/html; charset=utf-8"  />
<style>{{ template "styles" }}</style>
<title>Service Status</title>
<body>
<h1  style="text-align: center;">{{ .Code }} - {{ .Title }}  </h1>
</head>
</body>
</html>
```
</details>
<details><summary>style.css</summary>

```
{{ define "styles" }}
 .container {
   width: 100%;
   max-width: 100%;
 }
{{ end }}
```
</details>

---
## Template variables:
| Variable | Description |
| --- | --- |
| `{{.Code}}` | HTTP status code returned by the request |
| `{{.Title}}` |  HTTP status response message. [standard codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status) |
| `{{.Details.OriginalURI}}` | URI which originated the request |
| `{{.Details.ContentType}}` | Request Content-type |
| `{{.Details.Namespace}}` | Namespace where the backend Service is located |
| `{{.Details.IngressName}}` | Name of the Ingress where the backend is defined |
| `{{.Details.ServiceName}}` | Name of the Service backing the backend |
| `{{.Details.DeploymentName}}` | Name of the Deployment backend |
| `{{.Details.DesiredDeploymentReplicas}}` | Number of configured replicas in the backend deployment |
| `{{.Details.CurrentDeploymentReplicas}}` | Number of current running replicas for the backend deployment |
| `{{.Details.CustomFields.xxxx}}` | Custom fields defined in the service backing the backend (multiple custom fields are allowed)|

---
## Included templates:
 [**cniep**] : This template renders an error page with information about the current deployment state.

<p align="left">
  <img width="500px" src="https://raw.githubusercontent.com/marcportabellaclotet-mt/cniep/master/static/images/cniep-template.png"></img>
</p>

 [**default**] : Simple clean template.

---
## Build from source:

1 - # make build-all

2 - # docker build . -t cniep:xxx

