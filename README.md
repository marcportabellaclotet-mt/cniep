# Custom Nginx Ingress Error Pages (cniep)
-  [Description](#description)

-  [Installation](#installation)

-  [Enable CNIEP In K8s Services](#enable-cniep-for-k8s-services)

-  [Supported Annotations](#supported-annotations)

-  [How to template](#how-to-template)

-  [Supported template variables](#supported-template-variables)

-  [Build from source](#build-from-source)

-  [Bonus](#bonus)

---
## Description

**cniep** is a web service coupled to nginx ingress controller designed to provide customized error pages. It replaces the nginx ingress [default backend service](https://kubernetes.github.io/ingress-nginx/user-guide/default-backend/).
**cniep** traps pages returning error responses and displays a custom templated html page.
More info about about custom error codes can be found [here](https://kubernetes.github.io/ingress-nginx/user-guide/custom-errors/).
[ingress annotations](https://kubernetes.github.io/ingress-nginx/user-guide/nginx-configuration/annotations/#custom-http-errors) are required to trap error responses.
**cniep** is able to read k8s deployments information, and provides  variables that can be used in custom error response pages.

## Installation

1 - Check kubernetes folder, and customize the config to fit your setup.
2 - Deploy k8s **cniep** manifests:
``` 
# kubectl apply -f kubernetes 
```
3 - Modify your nginx controller to use cniep as default backend.
4 - Assuming that cniep is installed in cniep k8s namespace, add default-backend-service arg in your ingress nginx controller deployment.

```
      ...
      containers:
      - args:
        - /nginx-ingress-controller
        - --default-backend-service=cniep/cniep
      ...  
```

## Enable cniep for your k8s services
Steps to enable cniep for your k8s services:
1 - Install cniep deployment in your k8s server (check [installation](#installation) section)
2 - Annotate your k8s services. i.e:
```
apiVersion: v1
kind: Service
metadata:
  name: myservice
  annotations:
    cniep/deployment: myservice
    cniep/template: cniep
```
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

## Supported Annotations
* **cniep/template** (Required) : Defines the template used to render the error response pages.
* **cniep/deployment** (Optional) :  Defines the deployment backend for the annotated service. If this annotation is not set, it will search for service selectors ( app | app.kubernetes.io/name )
* **cniep/customjsonresponse-xxx** (Optional) : Use this annotation if you want **cniep** to render a custom json response instead of a templated html page. This annotation is error response code specific.
```
cniep/customjsonresponse-503: '{"json":"response"}'
```
 * **cniep/customjsonresponse-global** (Optional) : Use this annotation if you want **cniep** to render a custom json response instead of a templated html page. This annotation is for any error response code.
```
cniep/customjsonresponse-global: '{"json":"global-response"}'
```
* **cniep/s3-templatedir** (Optional) : Use this annotation if you prefer to upload the templates to an aws s3 bucket. **cniep** will sync templates from your s3 bucket to the running cniep pods. To use this feature, aws s3 credentials are required.
```
cniep/s3-templatedir: s3://mys3bucket/cniep2
cniep/template: cniep2
 ```
* **cniep/customfield-xxxx** (Optional) : Use this annotation if you want to pass custom fields to be rendered in your cniep templates. In the template section, there is more information on how to use these custom fields inside **cniep** templates.
```
 cniep/customfield-customvar: omg
 cniep/customfield-othervar: lol
```  

## How to template
**cniep** uses golang text/template package.
To render a custom error pages, **cniep** needs two type of files packaged inside a template folder:
```
mytemplate (template folder)
 - index.html (html file)
 - style.css (css style file)
```
index.html :
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
styles.css :
```
{{ define "styles" }}
 .container {
   width: 100%;
   max-width: 100%;
 }
{{ end }}
```
You can use specific pages per each status error code:
```
mytemplate (template folder)
 - index-503.html (html file)
 - style-503.css (css style file)
 - index.html (html file)
 - style.css (css style file)
```

## Supported template variables:
* **{{ .Code }}** : HTTP status code returned by the request
* **{{ .Title }}** :  HTTP status response message. i.e. Service Unavailable, Bad gateway ... [codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
* **{{ .Details.OriginalURI }}** : URI which originated the request.
* **{{ .Details.ContentType }}** : Content-type
* **{{ .Details.Namespace }}** : Namespace where the backend Service is located.
* **{{ .Details.IngressName }}** : Name of the Ingress where the backend is defined.
* **{{ .Details.ServiceName }}** : Name of the Service backing the backend.
* **{{ .Details.DeploymentName }}** : Name of the Deployment backend.
* **{{ .Details.DesiredDeploymentReplicas }}** : Number of configured replicas in the backend deployment.
* **{{ .Details.CurrentDeploymentReplicas }}** : Number of current running replicas for the backend deployment.
* **{{ .Details.CustomFields.xxxx }}** : Custom fields defined in the service backing the backend (multiple custom fields are allowed)

## Included templates:
 [**cniep**] : This template renders an error page with information about the current deployment state.

 [**ederm**] : This template renders an error page with information about the current deployment state and its AWS related DB. Requires ederm deployment to be installed.

 [**default**] : Simple clean template.

## Build from source:

1 - # make build-all

2 - # docker build . -t cniep:xxx

## Bonus

Use konami code in cniep template to get a cool bonus ;)
