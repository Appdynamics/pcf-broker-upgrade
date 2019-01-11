package main

import (
	"fmt"
	"github.com/cloudfoundry-community/go-cfclient"
	"io/ioutil"
	"net/url"
	"os"
)

const AppDServiceName = "appdynamics"
const UpgradeCSVFile = "appd-upgrade.csv"
const UpgradeScriptFile = "1_appd-upgrade.sh"
const ServiceBindingsScriptFile = "2_appd-bindings.sh"
const RestageAppsScriptFile = "3_appd-restage.sh"

type InstanceInfo struct {
	OrgName             string
	SpaceName           string
	ServiceName         string
	PlanName            string
	ServiceInstanceName string
	BoundApps           []string
}

func main() {
	config := &cfclient.Config{
		ApiAddress:        os.Getenv("CF_TARGET"),
		Username:          os.Getenv("CF_ADMIN_USERNAME"),
		Password:          os.Getenv("CF_ADMIN_PASSWORD"),
		SkipSslValidation: true,
	}

	fmt.Printf("Using configuration {%v:****} for CF Controller %v\n", config.Username, config.ApiAddress)

	client, err := cfclient.NewClient(config)
	if err != nil {
		fmt.Printf("unable to create cf client, exiting %v", err)
		return
	}

	instancesInfo, err := QueryInstanceInfo(client)
	if err != nil {
		fmt.Printf("Querying cf failed %v", err)
		return
	}

	if err := WriteCSVFile(instancesInfo); err != nil {
		fmt.Printf("Writing CSV file %s failed %v", UpgradeCSVFile, err)
	}

	if err := WriteUpgradeScript(instancesInfo); err != nil {
		fmt.Printf("Writing Upgrade script %s failed %v", UpgradeScriptFile, err)
	}

	if err := WriteBindScripts(instancesInfo); err != nil {
		fmt.Printf("Writing Binding scripts %s, %s  failed %v", ServiceBindingsScriptFile,
			RestageAppsScriptFile, err)
	}

	return
}

func WriteCSVFile(instanceInfos []InstanceInfo) error {
	csvString := ""
	for _, instance := range instanceInfos {
		csvString += fmt.Sprintf("%s,%s,%s,%s,%s\n", instance.OrgName, instance.SpaceName,
			instance.ServiceName, instance.PlanName, instance.ServiceInstanceName)
	}

	if err := ioutil.WriteFile(UpgradeCSVFile, []byte(csvString), 0644); err != nil {
		return err
	}

	return nil
}

func WriteUpgradeScript(instanceInfos []InstanceInfo) error {
	shString := ""
	for _, instance := range instanceInfos {
		shString += fmt.Sprintf("cf target -o %s -s %s\n", instance.OrgName, instance.SpaceName)
		shString += fmt.Sprintf("cf create-service %s %s %s\n", instance.ServiceName, instance.PlanName,
			instance.ServiceInstanceName)
	}

	if err := ioutil.WriteFile(UpgradeScriptFile, []byte(shString), 0755); err != nil {
		return err
	}

	return nil
}

func WriteBindScripts(instanceInfos []InstanceInfo) error {
	bindingString := ""
	restageString := ""

	for _, instance := range instanceInfos {
		boundApps := instance.BoundApps
		for _, app := range boundApps {
			bindingString += fmt.Sprintf("cf target -o %s -s %s\n", instance.OrgName, instance.SpaceName)
			bindingString += fmt.Sprintf("cf bind-service %s %s\n", app, instance.ServiceInstanceName)
			restageString += fmt.Sprintf("cf target -o %s -s %s\n", instance.OrgName, instance.SpaceName)
			restageString += fmt.Sprintf("cf restage %s\n", app)
		}
	}

	if err := ioutil.WriteFile(ServiceBindingsScriptFile, []byte(bindingString), 0755); err != nil {
		return err
	}

	if err := ioutil.WriteFile(RestageAppsScriptFile, []byte(restageString), 0755); err != nil {
		return err
	}
	return nil
}

func QueryInstanceInfo(client *cfclient.Client) ([]InstanceInfo, error) {
	sids, _ := client.ListServiceInstances()
	var infos []InstanceInfo

	for _, sid := range sids {
		if service, err := client.GetServiceByGuid(sid.ServiceGuid); err == nil && service.Label == AppDServiceName {
			plan, err := client.GetServicePlanByGUID(sid.ServicePlanGuid)
			if err != nil {
				fmt.Printf("%v", err)
			}
			space, err := client.GetSpaceByGuid(sid.SpaceGuid)
			if err != nil {
				fmt.Printf("%v", err)
			}
			org, err := space.Org()
			if err != nil {
				fmt.Printf("%v", err)
			}

			v := url.Values{}
			v.Set("q", fmt.Sprintf("service_instance_guid:%s", sid.Guid))
			bindings, err := client.ListServiceBindingsByQuery(v)
			if err != nil {
				fmt.Printf("%v", err)
			}

			fmt.Printf("writing info for %s, %s, %s, %s, %s\n", org.Name, space.Name, service.Label, plan.Name, sid.Name)

			var boundApps []string
			for _, binding := range bindings {
				app, _ := client.GetAppByGuid(binding.AppGuid)
				fmt.Printf("Binding found:  Application: %s - ServiceInstance: %s\n", app.Name, sid.Name)
				boundApps = append(boundApps, app.Name)
			}

			info := InstanceInfo{
				OrgName:             org.Name,
				SpaceName:           space.Name,
				ServiceName:         service.Label,
				PlanName:            plan.Name,
				ServiceInstanceName: sid.Name,
				BoundApps:           boundApps,
			}
			infos = append(infos, info)
		}
	}

	return infos, nil
}
