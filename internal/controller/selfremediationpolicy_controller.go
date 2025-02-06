/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	remediationv1alpha1 "kubemedic/api/v1alpha1"
)

// SelfRemediationPolicyReconciler reconciles a SelfRemediationPolicy object
type SelfRemediationPolicyReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	MetricsWatcher *MetricsWatcher
}

func NewSelfRemediationPolicyReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	metricsClient *versioned.Clientset,
	kubeClient *kubernetes.Clientset,
) *SelfRemediationPolicyReconciler {
	return &SelfRemediationPolicyReconciler{
		Client:         client,
		Scheme:         scheme,
		MetricsWatcher: NewMetricsWatcher(metricsClient, kubeClient),
	}
}

// Reconcile handles the reconciliation loop for SelfRemediationPolicy
func (r *SelfRemediationPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Get the policy
	var policy remediationv1alpha1.SelfRemediationPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Process each rule
	for _, rule := range policy.Spec.Rules {
		// Get target deployment
		var deployment appsv1.Deployment
		if err := r.Get(ctx, client.ObjectKey{
			Namespace: policy.Namespace,
			Name:      rule.Actions[0].Target.Name,
		}, &deployment); err != nil {
			log.Error(err, "failed to get deployment")
			continue
		}

		// Get pods for the deployment
		pods := &corev1.PodList{}
		if err := r.List(ctx, pods, client.InNamespace(policy.Namespace),
			client.MatchingLabels(deployment.Spec.Selector.MatchLabels)); err != nil {
			log.Error(err, "failed to list pods")
			continue
		}

		// Check each pod's metrics
		for _, pod := range pods.Items {
			for _, condition := range rule.Conditions {
				threshold, _ := strconv.ParseFloat(condition.Threshold, 64)
				duration, _ := time.ParseDuration(condition.Duration)

				isOverThreshold, err := r.MetricsWatcher.IsPodOverThreshold(
					ctx,
					pod.Namespace,
					pod.Name,
					threshold,
					duration,
				)
				if err != nil {
					log.Error(err, "failed to check pod metrics")
					continue
				}

				if isOverThreshold {
					// Execute remediation actions
					if err := r.executeActions(ctx, rule.Actions, &deployment); err != nil {
						log.Error(err, "failed to execute actions")
					}
				}
			}
		}
	}

	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

func (r *SelfRemediationPolicyReconciler) executeActions(
	ctx context.Context,
	actions []remediationv1alpha1.Action,
	deployment *appsv1.Deployment,
) error {
	for _, action := range actions {
		switch action.Type {
		case remediationv1alpha1.ScaleUp:
			if action.ScalingParams == nil || action.ScalingParams.TemporaryMaxReplicas == nil {
				return fmt.Errorf("scaling parameters required for ScaleUp action")
			}

			// Store original replicas for later reversion
			originalReplicas := deployment.Spec.Replicas
			if deployment.Annotations == nil {
				deployment.Annotations = make(map[string]string)
			}
			deployment.Annotations["kubemedic.io/original-replicas"] = fmt.Sprintf("%d", *originalReplicas)

			// Scale up
			newReplicas := *action.ScalingParams.TemporaryMaxReplicas
			deployment.Spec.Replicas = &newReplicas

			if err := r.Update(ctx, deployment); err != nil {
				return fmt.Errorf("failed to scale deployment: %v", err)
			}

			// Schedule reversion if duration is specified
			if action.ScalingParams.ScalingDuration != "" {
				duration, _ := time.ParseDuration(action.ScalingParams.ScalingDuration)
				go r.scheduleReversion(deployment, duration)
			}
		}
	}
	return nil
}

func (r *SelfRemediationPolicyReconciler) scheduleReversion(deployment *appsv1.Deployment, duration time.Duration) {
	time.Sleep(duration)

	ctx := context.Background()
	// Get the current deployment
	var currentDeployment appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: deployment.Namespace,
		Name:      deployment.Name,
	}, &currentDeployment); err != nil {
		return
	}

	// Get original replicas
	if originalStr, ok := currentDeployment.Annotations["kubemedic.io/original-replicas"]; ok {
		if original, err := strconv.ParseInt(originalStr, 10, 32); err == nil {
			originalReplicas := int32(original)
			currentDeployment.Spec.Replicas = &originalReplicas
			r.Update(ctx, &currentDeployment)
		}
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SelfRemediationPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&remediationv1alpha1.SelfRemediationPolicy{}).
		Complete(r)
}
