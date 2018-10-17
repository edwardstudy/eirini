package knative_test

import (
	"testing"

	"code.cloudfoundry.org/eirini"
	"code.cloudfoundry.org/eirini/models/cf"
	"code.cloudfoundry.org/eirini/opi"
	knife "github.com/julz/knife/pkg/knative"
	"github.com/knative/serving/pkg/apis/serving/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/api/core/v1"
)

func TestKnative(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Knative Suite")
}

func createLRP(processGUID, lastUpdated, routes string) *opi.LRP {
	return &opi.LRP{
		Name: processGUID,
		Command: []string{
			"/bin/sh",
			"-c",
			"while true; do echo hello; sleep 10;done",
		},
		Env: map[string]string{"env": "env-value"},
		// RunningInstances: 1,
		// TargetInstances:  2,
		Image: "busybox",
		Metadata: map[string]string{
			cf.ProcessGUID: processGUID,
			cf.LastUpdated: lastUpdated,
			cf.VcapAppUris: routes,
		},
	}
}

// FIXME - This function implements exactly the same as the source code. Implementing something twice to test it is suboptimal. Equivalent function can be found in k8s package
func toService(lrp *opi.LRP, namespace string) *v1alpha1.Service {
	envs := MapToEnvVar(lrp.Env)
	envs = append(envs, v1.EnvVar{
		Name: "POD_NAME",
		ValueFrom: &v1.EnvVarSource{
			FieldRef: &v1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	})

	service := knife.NewRunLatestService(eirini.GetInternalServiceName(lrp.Name),
		withRevisionTemplate(lrp.Image, lrp.Command, envs),
	)

	service.Namespace = namespace
	service.Labels = map[string]string{
		"name":   lrp.Name,
		"eirini": "eirini",
	}

	service.Annotations = map[string]string{
		eirini.RegisteredRoutes: lrp.Metadata[cf.VcapAppUris],
	}

	for k, v := range lrp.Metadata {
		service.Annotations[k] = v
	}

	return service
}

func withRevisionTemplate(image string, args []string, env []v1.EnvVar) knife.ConfigurationOption {
	return func(t *v1alpha1.ConfigurationSpec) {
		t.RevisionTemplate.Spec.Container.Image = image
		t.RevisionTemplate.Spec.Container.Args = args
		t.RevisionTemplate.Spec.Container.Env = env
	}
}

func MapToEnvVar(env map[string]string) []v1.EnvVar {
	envVars := []v1.EnvVar{}
	for k, v := range env {
		envVar := v1.EnvVar{Name: k, Value: v}
		envVars = append(envVars, envVar)
	}
	return envVars
}
