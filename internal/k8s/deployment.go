package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/henrywhitakercommify/restarter/internal/metrics"
	"github.com/prometheus/client_golang/prometheus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Deployment struct {
	client *kubernetes.Clientset

	name      string
	namespace string
}

func NewDeployment(c *kubernetes.Clientset, namespace, name string) *Deployment {
	return &Deployment{
		client:    c,
		name:      name,
		namespace: namespace,
	}
}

func (d *Deployment) Get(ctx context.Context) (*appsv1.Deployment, error) {
	return d.client.AppsV1().Deployments(d.namespace).Get(ctx, d.name, metav1.GetOptions{})
}

// Returns the percentage of pods that are ready
func (d *Deployment) Ready(ctx context.Context) (float64, error) {
	dep, err := d.Get(ctx)
	if err != nil {
		return 0, err
	}

	pods, err := d.client.CoreV1().Pods(d.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(&metav1.LabelSelector{
			MatchLabels: dep.Spec.Selector.MatchLabels,
		}),
	})
	if err != nil {
		return 0, fmt.Errorf("list pods for deployment: %w", err)
	}

	total := 0
	totalReady := 0
	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			total++
			ready := true
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status != corev1.ConditionTrue {
					ready = false
					break
				}
			}
			if ready {
				totalReady++
			}
		}
	}

	metrics.TotalPods.With(d.Labels()).Set(float64(total))
	metrics.ReadyPods.With(d.Labels()).Set(float64(totalReady))

	return (float64(totalReady) / float64(total)) * 100, nil
}

func (d *Deployment) Restart(ctx context.Context) error {
	deployment, err := d.Get(ctx)
	if err != nil {
		return err
	}

	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = make(map[string]string)
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().
		Format(time.RFC3339)

	if _, err := d.client.AppsV1().Deployments(d.namespace).Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
		return err
	}
	return nil
}
func (d Deployment) Labels() prometheus.Labels {
	return prometheus.Labels{"namespace": d.namespace, "deployment": d.name}
}
