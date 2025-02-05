package controller

import (
    "context"
    "fmt"
    "time"
    
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/metrics/pkg/client/clientset/versioned"
)

type MetricsWatcher struct {
    metricsClient *versioned.Clientset
}

func NewMetricsWatcher(metricsClient *versioned.Clientset) *MetricsWatcher {
    return &MetricsWatcher{
        metricsClient: metricsClient,
    }
}

func (w *MetricsWatcher) GetPodCPUUsage(ctx context.Context, namespace, podName string) (float64, error) {
    metrics, err := w.metricsClient.MetricsV1beta1().PodMetrics(namespace).Get(ctx, podName, metav1.GetOptions{})
    if err != nil {
        return 0, fmt.Errorf("failed to get pod metrics: %v", err)
    }

    var totalCPUUsage int64
    var totalCPULimit int64

    for _, container := range metrics.Containers {
        cpuUsage := container.Usage.Cpu().MilliValue()
        totalCPUUsage += cpuUsage
    }

    // Get the pod to find CPU limits
    pod, err := w.metricsClient.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
    if err != nil {
        return 0, fmt.Errorf("failed to get pod: %v", err)
    }

    for _, container := range pod.Spec.Containers {
        if container.Resources.Limits != nil {
            if cpu := container.Resources.Limits.Cpu(); cpu != nil {
                totalCPULimit += cpu.MilliValue()
            }
        }
    }

    if totalCPULimit == 0 {
        return 0, fmt.Errorf("no CPU limits set for pod")
    }

    return float64(totalCPUUsage) / float64(totalCPULimit) * 100, nil
}

func (w *MetricsWatcher) IsPodOverThreshold(ctx context.Context, namespace, podName string, threshold float64, duration time.Duration) (bool, error) {
    startTime := time.Now()
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return false, ctx.Err()
        case <-ticker.C:
            usage, err := w.GetPodCPUUsage(ctx, namespace, podName)
            if err != nil {
                return false, err
            }

            if usage < threshold {
                // Reset timer if usage drops below threshold
                startTime = time.Now()
                continue
            }

            if time.Since(startTime) >= duration {
                return true, nil
            }
        }
    }
} 