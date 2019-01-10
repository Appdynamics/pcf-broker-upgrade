# PCF Broker Update Tool

This tool assists in upgrading 1.x to 4.x AppDynamics Service Broker tile. 

As 1.x -> 4.x upgrades are not supported in PCF. The tool generates scripts that can be applied for recreating AppDynamics Service instances that have existed as before along with a seperate script to bind the service instances to the apps (if needed)

## Usage 

0. As a prerequisite you must have admin access to CF. You ust set the following env varibles for the tool to talk to cf controller. 

```
export CF_ADMIN_USERNAME=admin
export CF_ADMIN_PASSWORD=<pwd>
export CF_TARGET=https://apps.sys.pie-multi-az-blue.cfplatformeng.com
```

`CF_TARGET` is the api endpoint. You can find it by doing 

```
$ cf api
api endpoint:   https://api.sys.pie-multi-az-blue.cfplatformeng.com
```


1. Download the [tool](https://github.com/Appdynamics/pcf-broker-upgrade/releases)  Run the tool before deleting the 1.x tile 

```
$ ./appd-broker-upgrades 

writing info for appdynamics-org, appdynamics-space, appdynamics, appd454Controller, appd454
writing info for appdynamics-org, appdynamics-space, appdynamics, appdOther, appdOther
writing info for appdynamics-org, appdynamics-space, appdynamics, appdNoSSL, appd-python
writing info for appdynamics-org, appdynamics-space, appdynamics, appdNoSSL, appd
Binding found:  Application: cf-python - ServiceInstance: appd

```

2. The tool generates 4 files 
```
$ ls *appd*
1_appd-upgrade.sh	2_appd-bindings.sh	3_appd-restage.sh	appd-upgrade.csv
```

3. Verify CSV file to see if we have captured all the AppDynamics service instances. The csv file takes format `org-name,space-name,service-name,plan-name,instance-name`

```
$ cat appd-upgrade.csv 
appdynamics-org,appdynamics-space,appdynamics,appd454Controller,appd454
appdynamics-org,appdynamics-space,appdynamics,appdOther,appdOther
appdynamics-org,appdynamics-space,appdynamics,appdNoSSL,appd-python
appdynamics-org,appdynamics-space,appdynamics,appdNoSSL,appd
```

4. The other files are simple shell scripts that contain the commands to recreate the service instances and bindings(if needed). The files are prepended with numbers indicating the step numbers in the work flow `1_`, `2_`, `3_` 

```
$ cat 1_appd-upgrade.sh 
cf target -o appdynamics-org -s appdynamics-space
cf create-service appdynamics appd454Controller appd454
cf target -o appdynamics-org -s appdynamics-space
```

```
$ cat 2_appd-bindings.sh 
cf bind-service cf-python appd
```

```
$ cat 3_appd-restage.sh 
cf restage cf-python
```

5. Once the 1.x tile is uninstalled. Install 4.x tile and create plans with same names as before

6. Then apply scripts 1_appd-upgrade.sh, 2_appd-bindings.sh (if needed), 3_appd-restage.sh (if needed)


