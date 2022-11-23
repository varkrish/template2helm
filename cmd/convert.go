package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	dcconvert "github.com/csfreak/dc2deploy/pkg/convert"
	appsv1 "github.com/openshift/api/apps/v1"
	route "github.com/openshift/api/route/v1"
	template "github.com/openshift/api/template/v1"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"

	//"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

const (
	tplPathDefault                    = "."
	tplPathUsage                      = "Path to an OpenShift Template, relative or absolute"
	chartPathDefault                  = "."
	chartPathUsage                    = "Destination directory of the Chart."
	helmRepoPathDefault               = "https://pages.github.com"
	helmRepoPathUsage                 = "Helm common repository URL to add as dependency"
	helmRepoChartDependencyVerDefault = ">=1.1.0"
	helmRepoChatDependencyVerUsage    = "Provide the parent chart's version ex: 1.0.0"
	imageRegistryDefault              = "quay.io"
	imageRegistryUsage                = "Image Repository URL to be set in the chart for app images "
	helmCommonChartNameDefault        = "app-helm-common"
	helmCommChartNameUsage            = "Name of the dependent common chart. Default is " + helmCommonChartNameDefault
)

var (
	tplPath          string
	chartPath        string
	helmCommonUrl    string
	imageRepoBaseUrl string
	helmCommChartVer string
	helmCommonName   string
	convertCmd       = &cobra.Command{
		Use:   "convert",
		Short: "Given the path to an OpenShift template file, spit out a Helm chart.",
		Long:  `Long version...`,
		RunE: func(cmd *cobra.Command, args []string) error {

			var myTemplate template.Template
			yamlFile, err := ioutil.ReadFile(filepath.Clean(tplPath))
			if err != nil {
				return fmt.Errorf("Couldn't load template: %v", err)
			}

			// Convert to json first
			jsonB, err := yaml.YAMLToJSON(yamlFile)
			checkErr(err, fmt.Sprintf("Error trasnfoming yaml to json: \n%s", string(yamlFile)))

			err = json.Unmarshal(jsonB, &myTemplate)
			checkErr(err, "Unable to marshal template")

			// Convert myTemplate.Objects into individual files
			var templates []*chart.File
			values := make(map[string]interface{})
			var vals Values
			objectToTemplate(&myTemplate.Objects, &myTemplate.ObjectLabels, &templates, &vals)
			checkErr(err, "Failed object to template conversion")
			// Convert myTemplate.Parameters into a yaml string map
			err = paramsToValues(&myTemplate.Parameters, &values, &templates)
			checkErr(err, "Failed parameter to value conversion")
			valuesAsByte, err := yaml.Marshal(vals)
			checkErr(err, "Failed converting values to YAML")
			/* In our case, we  need to create the values.yaml, chart.yaml nad basic deployer*/
			var dependency chart.Dependency
			dependency = chart.Dependency{
				Name:       helmCommonName,
				Version:    helmCommChartVer,
				Repository: helmCommonUrl,
			}
			myChart := chart.Chart{
				Metadata: &chart.Metadata{
					Name:        myTemplate.ObjectMeta.Name,
					Version:     "1.0.0",
					Description: myTemplate.ObjectMeta.Annotations["description"],
					//Tags:        myTemplate.ObjectMeta.Annotations["tags"],
					Type:       "application",
					APIVersion: "v2",
				},
				Templates: templates,
				Values:    values,
				Raw:       []*chart.File{{Name: "values.yaml", Data: []byte(valuesAsByte)}},
			}
			myChart.Metadata.Dependencies = append(myChart.Metadata.Dependencies, &dependency)
			if myChart.Metadata.Name == "" {
				ext := filepath.Ext(tplPath)
				name := filepath.Base(string(tplPath))[0 : len(filepath.Base(string(tplPath)))-len(ext)]
				myChart.Metadata.Name = name
			}

			err = chartutil.SaveDir(&myChart, chartPath)
			checkErr(err, fmt.Sprintf("Failed to save chart %s", myChart.Metadata.Name))

			return nil
		},
	}
)

func init() {
	convertCmd.Flags().StringVarP(&tplPath, "template", "t", tplPathDefault, tplPathUsage)
	convertCmd.Flags().StringVarP(&chartPath, "chart", "c", chartPathDefault, chartPathUsage)
	convertCmd.Flags().StringVarP(&helmCommonName, "commonchartname", "n", helmCommonChartNameDefault, helmCommChartNameUsage)

	convertCmd.Flags().StringVarP(&helmCommonUrl, "helmrepo", "r", helmRepoPathDefault, helmRepoPathUsage)
	convertCmd.Flags().StringVarP(&helmCommChartVer, "helmdepver", "d", helmRepoChartDependencyVerDefault, helmRepoChatDependencyVerUsage)
	rootCmd.AddCommand(convertCmd)
}

func checkErr(err error, msg string) {
	if err != nil {
		fmt.Print(fmt.Errorf(msg + err.Error()))
		os.Exit(1)
	}
	return
}

func getVolumeVal(container *core.Container, values *Values, volumes *[]core.Volume) error {

	vals := *values
	oVolumes := *volumes
	volumeVals := make(map[string]Mount)
	oContainer := *container
	mount := Mount{}
	vals.VolumeSpec.Enabled = len(oVolumes) > 0
	if len(oVolumes) == 0 {
		return nil
	}
	for _, v := range oVolumes {
		if (v.ConfigMap != nil) && (v.ConfigMap.Name != "") {
			mount.Name = v.ConfigMap.Name
			mount.Ttype = "configmap"
			for _, k := range v.ConfigMap.Items {
				mount.Keys = append(mount.Keys, k.Key)
			}
		} else if (v.Secret != nil) && (v.Secret.SecretName != "") {
			mount.Name = v.Secret.SecretName
			mount.Ttype = "secret"
			for _, k := range v.Secret.Items {
				mount.Keys = append(mount.Keys, k.Key)
			}
		} else if (v.PersistentVolumeClaim != nil) && (v.PersistentVolumeClaim.ClaimName != "") {
			mount.Ttype = "pvc"
			mount.Name = v.PersistentVolumeClaim.ClaimName
		} else {
			log.Println("Volume specification %s is not supported", v.Name)
		}
		volumeVals[v.Name] = mount
	}

	for _, vMounts := range oContainer.VolumeMounts {

		mount = volumeVals[vMounts.Name]
		mount.MountPath = vMounts.MountPath
		vals.VolumeSpec.Mounts = append(vals.VolumeSpec.Mounts, mount)
	}
	*values = vals
	return nil
}

// Convert the object list in the openshift template to a set of template files in the chart
func objectToTemplate(objects *[]runtime.RawExtension, templateLabels *map[string]string, templates *[]*chart.File,
	values *Values) error {
	o := *objects

	m := make(map[string][]byte)
	seperator := []byte{'-', '-', '-', '\n'}
	vals := *values

	vals.Hpa.Enabled = true
	vals.Hpa.Maxreplicas = 2
	vals.Hpa.Targetcpuutilizationpercentage = 80
	vals.Image.Repository = imageRepoBaseUrl
	vals.Image.Tag = string(" ")
	//vals.Image.Namespace = ""
	vals.Pdb.Enabled = true
	vals.Pdb.Minavailable = 1
	vals.Strategy.Ttype = "RollingUpdate"
	vals.Strategy.MaxSurge = "100%"
	vals.Strategy.MaxUnavailable = "25%"
	vals.ServiceAccountOptions.Create = true
	vals.ServiceAccountOptions.Name = ""

	for _, v := range o {
		var k8sR unstructured.Unstructured
		err := json.Unmarshal(v.Raw, &k8sR)
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to unmarshal Raw resource\n%v\n", v.Raw) + err.Error())
		}

		labels := k8sR.GetLabels()
		if labels == nil {
			k8sR.SetLabels(*templateLabels)
		} else {
			for key, value := range *templateLabels {
				labels[key] = value
			}
			k8sR.SetLabels(labels)
		}
		updatedJSON, err := k8sR.MarshalJSON()
		if err != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to marshal Unstructured record to JSON\n%v\n", k8sR) + err.Error())
		}
		log.Printf("Creating a template for object %s", k8sR.GetKind())
		var data []byte
		var errh error
		data, errh = yaml.JSONToYAML(updatedJSON)
		if errh != nil {
			return fmt.Errorf(fmt.Sprintf("Failed to marshal Raw resource back to YAML\n%v\n", updatedJSON) + errh.Error())
		}
		//var consolidatedEnv []byte
		var deploy = &v1.Deployment{}
		if k8sR.GetKind() == "DeploymentConfig" {
			dc := &appsv1.DeploymentConfig{}
			fmt.Printf("Converting DC to Deployment\n")
			errh = json.Unmarshal(updatedJSON, &dc)
			checkErr(err, "Unable to marshal template")
			deploy, errh = dcconvert.ToDeploy(dc)
			if err != nil {
				return fmt.Errorf("unable to convert to deploy: %w", err)
			}

			fmt.Printf("Successfully converted DC to Deployment\n")

			k8sR.SetKind(deploy.Kind)
		}

		if k8sR.GetKind() == "Deployment" {
			errh = json.Unmarshal(updatedJSON, &deploy)
			vals.Env = deploy.Spec.Template.Spec.Containers[0].Env
			vals.Resources = deploy.Spec.Template.Spec.Containers[0].Resources
			if deploy.Spec.Template.Spec.Containers[0].ReadinessProbe != nil {
				vals.Probes.Readiness = deploy.Spec.Template.Spec.Containers[0].ReadinessProbe
			}
			if deploy.Spec.Template.Spec.Containers[0].LivenessProbe != nil {
				vals.Probes.Liveness = deploy.Spec.Template.Spec.Containers[0].LivenessProbe
			}
			vals.Image.PullPolicy = deploy.Spec.Template.Spec.Containers[0].ImagePullPolicy
			vals.Image.Name = deploy.ObjectMeta.Name
			vals.Controller.Enabled = true
			vals.Controller.Type = "deployment"
			vals.Replicas.Min = *deploy.Spec.Replicas
			vals.Replicas.Max = *deploy.Spec.Replicas

			vals.Configs.Secrets = make([]string, 0)
			vals.Configs.Configmaps = make([]string, 0)

			for _, envFrom := range deploy.Spec.Template.Spec.Containers[0].EnvFrom {

				if envFrom.SecretRef != nil && envFrom.SecretRef.Name != "" {
					vals.Configs.Secrets = append(vals.Configs.Secrets, envFrom.SecretRef.Name)
				}
				if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name != "" {
					vals.Configs.Configmaps = append(vals.Configs.Configmaps, envFrom.ConfigMapRef.Name)
				}
			}
			err := getVolumeVal(&deploy.Spec.Template.Spec.Containers[0], &vals, &deploy.Spec.Template.Spec.Volumes)
			if err != nil {
				return fmt.Errorf("unable to convert to deploy: %w", err)
			}
		}

		if k8sR.GetKind() == "Service" {
			var svc = &core.Service{}
			log.Println("Converting yaml to Service")
			errh = json.Unmarshal(updatedJSON, &svc)
			checkErr(err, "Unable to marshal template")
			log.Println("Converted  yaml to Service")
			vals.Service.Annotations = svc.Annotations
			vals.Service.Enabled = true
			vals.Service.Svc_type = string(svc.Spec.Type)
			vals.Service.Ports = make([]string, len(svc.Spec.Ports))
			for i, v := range svc.Spec.Ports {
				vals.Service.Ports[i] = fmt.Sprint(v.Port)
			}
		}
		if k8sR.GetKind() == "Route" {
			var route = route.Route{}
			log.Println("Converting yaml to Service")
			errh = json.Unmarshal(updatedJSON, &route)
			checkErr(err, "Unable to marshal template")
			log.Println("Converted  yaml to Service")
			vals.Route.Annotations = route.Annotations
			vals.Route.Enabled = true
			vals.Route.Hostname = route.Spec.Host
		}
		if k8sR.GetKind() != "Deployment" &&
			k8sR.GetKind() != "Job" &&
			k8sR.GetKind() != "CronJob" &&
			k8sR.GetKind() != "ImageStream" &&
			k8sR.GetKind() != "Route" &&
			k8sR.GetKind() != "Service" {
			if m[k8sR.GetKind()] == nil {
				m[k8sR.GetKind()] = data

			} else {
				newdata := append(m[k8sR.GetKind()], seperator...)
				data = append(newdata, data...)
				m[k8sR.GetKind()] = data
			}
		}
	}
	//Create chart using map
	for k, v := range m {
		fmt.Printf("----%s----", k)
		fmt.Print(k != "Deployment")

		if (k != "Deployment") && (k != "Job") &&
			(k != "CronJob") {
			name := "templates/" + strings.ToLower(k+".yaml")
			tf := chart.File{
				Name: name,
				Data: v,
			}
			*templates = append(*templates, &tf)
		}
	}
	*values = vals
	return nil
}

func paramsToValues(param *[]template.Parameter, values *map[string]interface{}, templates *[]*chart.File) error {

	p := *param
	t := *templates
	v := *values

	for _, pm := range p {
		name := strings.ToLower(pm.Name)
		log.Printf("Convert parameter %s to value .%s", pm.Name, name)

		for i, tf := range t {
			// Search and replace ${PARAM} with {{ .Values.param }}
			raw := tf.Data
			// Handle string format parameters
			ns := strings.ReplaceAll(string(raw), fmt.Sprintf("${%s}", pm.Name), fmt.Sprintf("{{ .Values.%s }}", name))
			// TODO Handle binary formatted data differently
			ns = strings.ReplaceAll(ns, fmt.Sprintf("${{%s}}", pm.Name), fmt.Sprintf("{{ .Values.%s }}", name))
			ntf := chart.File{
				Name: tf.Name,
				Data: []byte(ns),
			}

			t[i] = &ntf
		}

		if pm.Value != "" {
			v[name] = pm.Value
		} else {
			v[name] = "# TODO: must define a default value for ." + name
		}
	}
	ntf := chart.File{
		Name: "templates/basic_deployment.yaml",
		Data: []byte("{{- include \"common.appSpec\" . -}}"),
	}
	t = append(t, &ntf)
	*templates = t
	*values = v

	return nil
}
