package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cloudfoundry-community/go-cfenv"
	"github.com/cloudfoundry-community/types-cf"
	"github.com/go-martini/martini"
)

type serviceInstanceResponse struct {
	DashboardURL string `json:"dashboard_url"`
}

type serviceBindingResponse struct {
	Credentials    map[string]interface{} `json:"credentials"`
	SyslogDrainURL string                 `json:"syslog_drain_url,omitempty"`
}

var serviceName, servicePlan string
var serviceBinding serviceBindingResponse
var appURL string

func brokerCatalog() (int, []byte) {
	catalog := cf.Catalog{
		Services: []*cf.Service{
			{
				ID:          "-service-" + serviceName,
				Name:        serviceName,
				Description: "Shared service for " + serviceName,
				Bindable:    true,
				Metadata: &cf.ServiceMeta{
					DisplayName: serviceName,
				},
				Plans: []*cf.Plan{
					{
						ID:          "plan-" + servicePlan,
						Name:        servicePlan,
						Description: "Shared service for " + serviceName,
						Free:        true,
					},
				},
			},
		},
	}
	json, err := json.Marshal(catalog)
	if err != nil {
		return 500, []byte{}
	}
	return 200, json
}

func createServiceInstance(params martini.Params) (int, []byte) {
	instance := serviceInstanceResponse{DashboardURL: fmt.Sprintf("%s/dashboard", appURL)}
	json, err := json.Marshal(instance)
	if err != nil {
		return 500, []byte{}
	}
	return 201, json
}

func deleteServiceInstance(params martini.Params) (int, string) {
	return 200, "{}"
}

func createServiceBinding(params martini.Params) (int, []byte) {
	json, err := json.Marshal(serviceBinding)
	if err != nil {
		return 500, []byte{}
	}
	return 201, json
}

func deleteServiceBinding(params martini.Params) (int, string) {
	return 200, "{}"
}

func showServiceInstanceDashboard(params martini.Params) (int, string) {
	return 200, "Dashboard"
}

func main() {
	m := martini.Classic()

	serviceName = os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "some-service-name" // replace with cfenv.AppName
	}
	servicePlan = os.Getenv("SERVICE_PLAN")
	if servicePlan == "" {
		servicePlan = "shared"
	}

	credentials := os.Getenv("CREDENTIALS")
	if credentials == "" {
		credentials = "{\"port\": \"4000\"}"
	}

	json.Unmarshal([]byte(credentials), &serviceBinding.Credentials)

	appEnv, err := cfenv.Current()
	if err != nil {
		panic(err)
	}
	appURL = fmt.Sprintf("http://%s", appEnv.ApplicationURIs[0])

	// Cloud Foundry Service API
	m.Get("/v2/catalog", brokerCatalog)
	m.Put("/v2/service_instances/:service_id", createServiceInstance)
	m.Delete("/v2/service_instances/:service_id", deleteServiceInstance)
	m.Put("/v2/service_instances/:service_id/service_bindings/:binding_id", createServiceBinding)
	m.Delete("/v2/service_instances/:service_id/service_bindings/:binding_id", deleteServiceBinding)

	// Service Instance Dashboard
	m.Get("/dashboard", showServiceInstanceDashboard)

	m.Run()
}
