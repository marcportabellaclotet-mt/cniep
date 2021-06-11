package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreClient "k8s.io/client-go/kubernetes"
	restClient "k8s.io/client-go/rest"
	cmdClient "k8s.io/client-go/tools/clientcmd"
)

const (
	svcScanInterval = 10
)

var serviceDetailsMap = map[string]serviceDetails{}

type serviceDetails struct {
	annotations         interface{}
	backend             string
	deployName          string
	customErrorTemplate string
	customJsonResponse  map[string]string
	customFields        map[string]string
	desiredReplicas     int32
	currentReplicas     int32
}

type K8s struct {
	Clientset  coreClient.Interface
	RestConfig *restClient.Config
}

func NewK8s() (*K8s, error) {
	client := K8s{}
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster == true {
		config, err := restClient.InClusterConfig()
		client.RestConfig = config
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
		client.Clientset, err = coreClient.NewForConfig(config)
		if err != nil {
			logrus.Error(err.Error())
			return nil, err
		}
		return &client, nil
	}
	kubeconfig := fmt.Sprintf("%v/.kube/config", homeDir())
	config, err := cmdClient.BuildConfigFromFlags("", kubeconfig)
	client.RestConfig = config
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	client.Clientset, err = coreClient.NewForConfig(config)
	if err != nil {
		logrus.Error(err.Error())
		return nil, err
	}
	return &client, nil
}

func getDeploymentReplicas(deployment string, namespace string) (desiredReplicacount int32, currentReplicacount int32, err error) {
	client, err := NewK8s()
	if err != nil {
		logrus.Error(err.Error())
		return -1, -1, err
	}
	deploymentClient := client.Clientset.AppsV1().Deployments(namespace)
	deploymentInfo, err := deploymentClient.Get(context.TODO(), deployment, metav1.GetOptions{})
	if err != nil {
		logrus.Error(err.Error())
		return -1, -1, err
	}
	desiredReplicacount = *deploymentInfo.Spec.Replicas
	currentReplicacount = deploymentInfo.Status.ReadyReplicas
	return desiredReplicacount, currentReplicacount, err
}

func getServiceDetails(serviceName string, namespace string) (serviceDetails serviceDetails, err error) {
	client, err := NewK8s()
	if err != nil {
		logrus.Error(err.Error())
		return serviceDetails, err
	}
	serviceClient := client.Clientset.CoreV1().Services(namespace)
	serviceInfo, err := serviceClient.Get(context.TODO(), serviceName, metav1.GetOptions{})
	if err != nil {
		logrus.Error(err.Error())
		return serviceDetails, err
	}
	annotationsInfo, err := InterfaceMap(serviceInfo.Annotations)
	if err != nil {
		logrus.Error(err.Error())
		return serviceDetails, err
	}
	serviceDetails.annotations = annotationsInfo
	for k, v := range serviceInfo.Annotations {
		switch {
		case k == "cniep/s3-templatedir":
			listS3Objects(v)
		case k == "cniep/deployment":
			serviceDetails.deployName = v
		case k == "cniep/template":
			serviceDetails.customErrorTemplate = v
		case strings.HasPrefix(k, "cniep/customjsonresponse-"):
			if serviceDetails.customJsonResponse == nil {
				serviceDetails.customJsonResponse = make(map[string]string)
			}
			customCode := strings.Split(k, "cniep/customjsonresponse-")[1]
			serviceDetails.customJsonResponse[customCode] = v
		case strings.HasPrefix(k, "cniep/customfield-"):
			if serviceDetails.customFields == nil {
				serviceDetails.customFields = make(map[string]string)
			}
			customField := strings.Split(k, "cniep/customfield-")[1]
			serviceDetails.customFields[customField] = v
		}
	}
	if serviceDetails.deployName == "" {
		if val, ok := serviceInfo.Spec.Selector["app"]; ok {
			serviceDetails.deployName = val
		} else if val, ok := serviceInfo.Spec.Selector["app.kubernetes.io/name"]; ok {
			serviceDetails.deployName = val
		} else {
			serviceDetails.deployName = "unknown"
		}
	}
	desiredReplicas, currentReplicas, err := getDeploymentReplicas(serviceDetails.deployName, namespace)
	if err != nil {
		logrus.Error(err.Error())
		return serviceDetails, err
	}
	serviceDetails.desiredReplicas = desiredReplicas
	serviceDetails.currentReplicas = currentReplicas
	return serviceDetails, err
}

func scanServices() (err error) {
	logrus.Info("Starting k8s service scan")
	for {
		client, err := NewK8s()
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		servicesClient := client.Clientset.CoreV1().Services(v1.NamespaceAll)
		list, err := servicesClient.List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			logrus.Error(err.Error())
			return err
		}
		for _, service := range list.Items {
			for k := range service.Annotations {
				if strings.HasPrefix(k, "cniep") {
					mapIdentifier := fmt.Sprintf("%v.%v", service.Name, service.Namespace)
					serviceDetailsMap[mapIdentifier], err = getServiceDetails(service.Name, service.Namespace)
					break
				}
			}
		}
		time.Sleep(svcScanInterval * time.Second)
	}
}
