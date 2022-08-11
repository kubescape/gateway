package main

import (
	"flag"
	"os"

	"github.com/kubescape/gateway/pkg"

	"github.com/armosec/utils-k8s-go/probes"
	logger "github.com/dwertent/go-logger"
	"github.com/dwertent/go-logger/helpers"
)

//go:generate swagger generate spec -o ./docs/swagger.yaml
func main() {
	flag.Parse()

	displayBuildTag()

	isReadinessReady := false
	go probes.InitReadinessV1(&isReadinessReady)

	gateway := gateway.NewGateway()
	isReadinessReady = true
	gateway.SetupAndServe()
}

// DisplayBuildTag outputs the bulid tag of the current release
func displayBuildTag() {
	logger.L().Info("Image version", helpers.String("release", os.Getenv(gateway.ReleaseBuildTagEnvironmentVariable)))
}
