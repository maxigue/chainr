package worker

import "testing"

func TestMakeK8SJob(t *testing.T) {
	cp := K8SCloudProvider{}

	job := Job{"test", "busybox", "exit 0"}
	k8sJob := cp.makeK8SJob(job)

	container := k8sJob.Spec.Template.Spec.Containers[0]
	if container.Name != "test" {
		t.Errorf("container.Name = %v, expected test", container.Name)
	}
	if container.Image != "busybox" {
		t.Errorf("container.Image = %v, expected busybox", container.Image)
	}

	expectedCommand := []string{"sh", "-c", "exit 0"}
	for i := range container.Command {
		if container.Command[i] != expectedCommand[i] {
			t.Errorf("container.Command = %v, expected %v", container.Command, expectedCommand)
		}
	}
}
