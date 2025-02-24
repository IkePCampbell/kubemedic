package controller

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

type MetricsWatcher struct {
	metricsClient versioned.Interface
	kubeClient    *kubernetes.Clientset
}

func NewMetricsWatcher(metricsClient versioned.Interface) *MetricsWatcher {
	if metricsClient == nil {
		return nil
	}
	return &MetricsWatcher{
		metricsClient: metricsClient,
	}
}

func (w *MetricsWatcher) GetPodCPUUsage(pod *corev1.Pod) (float64, error) {
	if w == nil {
		return 0, fmt.Errorf("metrics watcher is nil")
	}
	if pod == nil {
		return 0, fmt.Errorf("pod is nil")
	}
	if w.metricsClient == nil {
		return 0, fmt.Errorf("metrics client not initialized")
	}

	metrics, err := w.metricsClient.MetricsV1beta1().PodMetricses(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get pod metrics: %w", err)
	}

	if metrics == nil || len(metrics.Containers) == 0 {
		return 0, fmt.Errorf("no metrics data available")
	}

	var totalCPU float64
	for _, container := range metrics.Containers {
		cpuQuantity := container.Usage.Cpu()
		if cpuQuantity != nil {
			totalCPU += float64(cpuQuantity.MilliValue()) / 1000.0
		}
	}

	return totalCPU, nil
}

func (w *MetricsWatcher) IsPodOverThreshold(pod *corev1.Pod, threshold float64) (bool, error) {
	if pod == nil {
		return false, fmt.Errorf("pod is nil")
	}

	cpuUsage, err := w.GetPodCPUUsage(pod)
	if err != nil {
		return false, fmt.Errorf("failed to get CPU usage: %w", err)
	}

	return cpuUsage > threshold, nil
}
