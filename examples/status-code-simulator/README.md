# Testing cniep responses

### Deploy the cniep status-code-simulator
```
# kubectl apply -f status-code-simulator
```
ingress has to be properly configured. Change the hostname to match your setup.

1 - Navigate to http://cniep.example.com/503/ I
 
A custom cniep error page will be rendered based on cniep template. Status respone code is 503.

2 - Navigate to http://cniep.example.com/502/

A blank page will be rendered. Status respone code is 502 (cniep-test ingress is configured to trap only 503 error)

3 - Edit cniep-test service and change cniep annotations

``` 
# kubectl edit svc cniep-test
``` 

4 -change annotation [ cniep/template: cniep ] to [ cniep/template: default ]

5 - Navigate to Navigate to http://cniep.example.com/503/
In a few seconds the error response page will be switched to a basic html page.

6 - Edit cniep-test service again
``` 
# kubectl edit svc cniep-test
``` 

8 - Add cniep annotations : cniep/customjsonresponse-503: '{"cniep":"response"}'
In a few seconds the error response page will be switched to json data.

9 - Do your own tests. Refer to Main README.md to check all possible configurations
