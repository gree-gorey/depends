package main

import (
	"fmt"
	"time"
	"flag"
	"strconv"
	"io/ioutil"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	namespaceFile, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
  if err != nil {
      fmt.Print(err)
  }
	namespace := string(namespaceFile)

	interval := flag.Int("interval", 2, "Interval in seconds")
	services := flag.String("services", "redis:1", "Services to wait for, separated by comma")
	flag.Parse()

	servicesList := strings.Split(*services, ",")
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	allReady := false
	for allReady != true {
		allReady = true
		fmt.Printf("Iteration...\n")
		for _, serviceName := range servicesList {

			addressesCount := 1
			serviceNameList := strings.Split(serviceName, ":")
			service := serviceNameList[0]
			if len(serviceNameList) > 1 {
				addressesCount, _ = strconv.Atoi(serviceNameList[1])
			}

			ep, err := clientset.CoreV1().Endpoints(namespace).Get(service, metav1.GetOptions{})
			if errors.IsNotFound(err) {
				fmt.Printf("  Endpoint not found\n")
				allReady = false
			} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
				fmt.Printf("  Error getting endpoint %v\n", statusError.ErrStatus.Message)
				allReady = false
			} else if err != nil {
				panic(err.Error())
				allReady = false
			} else {
				notReadyCount := 0
				readyCount := 0
				readyAddressesCount := 0
				subsetsCount := len(ep.Subsets)
				for _, subset := range ep.Subsets {
					if subset.Addresses != nil {
					  readyCount++
						readyAddressesCount += len(subset.Addresses)
					}
					if subset.NotReadyAddresses != nil {
						notReadyCount++
					}
				}
				fmt.Printf("  [*] Found %d subsets of service \"%s\"\n", len(ep.Subsets), service)
				fmt.Printf("   |    %d of them has ready addresses\n", readyCount)
				fmt.Printf("   |    %d of them has NOT ready addresses\n", notReadyCount)
				fmt.Printf("   |    count of ready addresses: %d\n", readyAddressesCount)

				currentReady := (notReadyCount == 0) && (readyCount == subsetsCount) && (subsetsCount > 0) && (readyAddressesCount >= addressesCount)

				if currentReady != true {
					fmt.Printf("   |  Service \"%s\" is NOT ready\n", service)
					allReady = false
				} else {
					fmt.Printf("   |  Service \"%s\" is ready\n", service)
				}
			}
		}
		if allReady {
			fmt.Printf("Dependencies are ready! Quitting\n")
		}
		time.Sleep(time.Duration(*interval) * time.Second)
	}
}
