package component

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestInitContainer(t *testing.T) {

	initContainer := &corev1.Container{
		Name:            "test-container",
		Image:           "busybox",
		ImagePullPolicy: corev1.PullAlways,
		Resources:       corev1.ResourceRequirements{},
		Env:             []corev1.EnvVar{},
		Command:         []string{"/bin/sh", "-c"},
		Args:            []string{"while true; do sleep 1; if [ -f " + completionFile + " ]; then break; fi done"},
		VolumeMounts: []corev1.VolumeMount{
			buildContextVolumeMount,
		},
	}

	initContainerPorted := corev1.Container{}

	want := initContainer
	got := InitContainer("test-container")
	initContainerPorted = *got

	if !reflect.DeepEqual(initContainerPorted, *want) {
		t.Errorf("Container parameters do not match")
	}
}
