package main

import (
	"errors"
	"fmt"
	"github.com/nabsul/k8s-ecr-login-renew/src/aws"
	"github.com/nabsul/k8s-ecr-login-renew/src/k8s"
	"os"
	"strings"
	"time"
)

const (
	envVarAwsSecret       = "DOCKER_SECRET_NAME"
	envVarTargetNamespace = "TARGET_NAMESPACE"
	envVarRegistries      = "DOCKER_REGISTRIES"
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("Running at " + time.Now().UTC().String())

	name := os.Getenv(envVarAwsSecret)
	if name == "" {
		panic(fmt.Sprintf("Environment variable %s is required", name))
	}

	namespaceList := os.Getenv(envVarTargetNamespace)
	if namespaceList == "" {
		namespaceList = "default"
	}

	fmt.Print("Fetching auth data from AWS... ")
	username, password, server, err := aws.GetUserAndPass()
	checkErr(err)
	fmt.Println("Success.")

	servers := []string{server}
	registries := strings.Split(os.Getenv(envVarRegistries), ",")
	for _, registry := range registries {
		servers = append(servers, registry)
	}

	namespaces, err := k8s.GetNamespaces(namespaceList)
	checkErr(err)
	fmt.Printf("Updating kubernetes secret [%s] in %d namespaces\n", name, len(namespaces))

	failed := false
	for _, ns := range namespaces {
		fmt.Printf("Updating secret in namespace [%s]... ", ns)
		err = k8s.UpdatePassword(ns, name, username, password, servers)
		if nil != err {
			fmt.Printf("failed: %s\n", err)
			failed = true
		} else {
			fmt.Println("success")
		}
	}

	if failed {
		panic(errors.New("failed to create one of more Docker login secrets"))
	}

	fmt.Println("Job complete.")
}
