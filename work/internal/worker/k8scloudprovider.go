package worker

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8SCloudProvider struct {
	kube      kubernetes.Interface
	namespace string
}

// If the Kubernetes client can not be created,
// this function panics.
func NewK8SCloudProvider() K8SCloudProvider {
	var cp K8SCloudProvider

	if val, ok := os.LookupEnv("KUBECONFIG"); ok {
		log.Println("Program is running outside the cluster, loading config from", val)
		cp = newOutsideCluster(val)
	} else {
		log.Println("Program is running inside the cluster")
		cp = newInsideCluster()
	}

	log.Println("Jobs will be run on namespace", cp.namespace)
	return cp
}

func newOutsideCluster(kubeconfig string) K8SCloudProvider {
	loadingRules := clientcmd.ClientConfigLoadingRules{
		ExplicitPath: kubeconfig,
	}
	config, err := loadingRules.Load()
	if err != nil {
		panic(err)
	}
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewDefaultClientConfig(*config, configOverrides)

	clientConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		panic(err)
	}

	clientset := kubernetes.NewForConfigOrDie(clientConfig)
	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		panic(err)
	}

	return K8SCloudProvider{clientset, namespace}
}

func newInsideCluster() K8SCloudProvider {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset := kubernetes.NewForConfigOrDie(config)

	namespaceFile := "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	data, err := ioutil.ReadFile(namespaceFile)
	if err != nil {
		panic("namespace file could not be read at " + namespaceFile)
	}
	namespace := strings.TrimSpace(string(data))

	return K8SCloudProvider{clientset, namespace}
}

func (cp K8SCloudProvider) RunJob(job Job) error {
	k8sJob := cp.makeK8SJob(job)
	created, err := cp.kube.BatchV1().Jobs(cp.namespace).Create(&k8sJob)
	if err != nil {
		return err
	}
	defer cp.deleteK8SJob(created.Name)

	watch, err := cp.kube.BatchV1().Jobs(cp.namespace).Watch(metav1.ListOptions{
		FieldSelector: fields.Set{
			"metadata.name": created.Name,
		}.AsSelector().String(),
	})
	if err != nil {
		return err
	}

	for event := range watch.ResultChan() {
		j, ok := event.Object.(*batchv1.Job)
		if !ok {
			return errors.New("unexpected type")
		}
		if j.Status.Failed > 0 {
			return errors.New("job execution failed")
		} else if j.Status.Succeeded > 0 {
			return nil
		}
	}

	return nil
}

func (cp K8SCloudProvider) makeK8SJob(job Job) batchv1.Job {
	var k8sJob batchv1.Job
	k8sJob.GenerateName = "chainr-job-"
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "chainr",
	}
	k8sJob.Labels = labels
	var backoffLimit int32 = 0
	k8sJob.Spec.BackoffLimit = &backoffLimit
	k8sJob.Spec.Template.Labels = labels
	k8sJob.Spec.Template.Spec.Containers = []corev1.Container{
		corev1.Container{
			Name:    job.Name,
			Image:   job.Image,
			Command: []string{"sh", "-c", job.Run},
		},
	}
	k8sJob.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever

	return k8sJob
}

func (cp K8SCloudProvider) deleteK8SJob(name string) {
	propagationPolicy := metav1.DeletePropagationForeground
	opts := metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	}

	if err := cp.kube.BatchV1().Jobs(cp.namespace).Delete(name, &opts); err != nil {
		log.Println("Unable to delete Kubernetes job", name)
	}
}
