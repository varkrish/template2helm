/*
Copyright © 2022 Jason Ross

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package cmd

import (
	route "github.com/openshift/api/route/v1"
	core "k8s.io/api/core/v1"
	//"github.com/spf13/cobra"
)

const (
	LastAppliedAnnotationKey = "kubectl.kubernetes.io/last-applied-configuration"
	GeneratedByAnnotationKey = "openshift.io/generated-by"
	DeploymentConfigPodLabel = "deploymentconfig"
	DeploymentPodLabel       = "deployment"
)

var (
	StripAnnotations = [2]string{LastAppliedAnnotationKey, GeneratedByAnnotationKey}
	ReplaceLabels    = map[string]string{
		DeploymentConfigPodLabel: DeploymentPodLabel,
	}
)

type Mount struct {
	Name        string   `json:"name,omitempty"`
	Ttype       string   `json:"type,omitempty"`
	MountPath   string   `json:"mountPath,omitempty"`
	Keys        []string `json:"keys,omitempty"`
	DefaultMode string   `json:"defaultMode,omitempty"`
}
type VolumeSpec struct {
	Enabled bool    `json:"enabled,omitempty"`
	Mounts  []Mount `json:"mounts,omitempty"`
}
type Controller struct {
	Enabled bool   `json:"enabled,omitempty"`
	Type    string `json:"type,omitempty"`
}
type Replicas struct {
	Min int32 `json:"min,omitempty"`
	Max int32 `json:"max,omitempty"`
}
type Configs struct {
	Configmaps []string `json:"configmaps,omitempty"`
	Secrets    []string `json:"secrets,omitempty"`
}
type Hpa struct {
	Enabled                        bool `json:"enabled,omitempty"`
	Maxreplicas                    int  `json:"maxreplicas,omitempty"`
	Minreplicas                    int  `json:"minreplicas,omitempty"`
	Targetcpuutilizationpercentage int  `json:"targetcpuutilizationpercentage,omitempty"`
}
type Pdb struct {
	Enabled      bool `json:"enabled,omitempty"`
	Minavailable int  `json:"minavailable,omitempty"`
}
type probes struct {
	Liveness  *core.Probe `json:"liveness,omitempty"`
	Readiness *core.Probe `json:"readiness,omitempty"`
}
type Route struct {
	Enabled     bool              `json:"enabled,omitempty"`
	Hostname    string            `json:"hostname,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	TLSConfig   route.TLSConfig   `json:"termination"`
}
type Service struct {
	Enabled     bool              `json:"enabled,omitempty"`
	Svc_type    string            `json:"type,omitempty"`
	Ports       []string          `json:"ports,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}
type ImageSpec struct {
	Repository string          `json:"repository,omitempty"`
	Namespace  string          `json:"namespace,omitempty"`
	Name       string          `json:"name,omitempty"`
	Tag        string          `json:"tag,omitempty"`
	PullPolicy core.PullPolicy `json:"pullPolicy,omitempty"`
}
type Values struct {
	Replicacount int                       `json:"replicacount,omitempty"`
	Env          []core.EnvVar             `json:"env,omitempty"`
	Image        ImageSpec                 `json:"image,omitempty"`
	Service      Service                   `json:"service,omitempty"`
	Resources    core.ResourceRequirements `json:"resources,omitempty"`
	Probes       probes                    `json:"probes,omitempty"`
	Controller   Controller                `json:"controller,omitempty"`
	Route        Route                     `json:"route,omitempty"`
	VolumeSpec   VolumeSpec                `json:"volumeMounts,omitempty"`
	Replicas     Replicas                  `json:"replicas,omitempty"`
	Configs      Configs                   `json:"configs,omitempty"`
	Hpa          Hpa                       `json:"hpa,omitempty"`
	Pdb          Pdb                       `json:"pdb,omitempty"`
}
