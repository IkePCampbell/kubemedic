/*
Copyright 2024.

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
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	remediationv1alpha1 "kubemedic/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SelfRemediationPolicyReconciler reconciles a SelfRemediationPolicy object
type SelfRemediationPolicyReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	MetricsWatcher *MetricsWatcher
	// Track active remediations
	activeRemediations sync.Map
}

// RemediationState tracks the state of active remediations
type RemediationState struct {
	LastChecked time.Time
	Policy      types.NamespacedName
	Target      types.NamespacedName
}

func NewSelfRemediationPolicyReconciler(
	client client.Client,
	scheme *runtime.Scheme,
	metricsClient versioned.Interface,
) *SelfRemediationPolicyReconciler {
	if client == nil {
		panic("client cannot be nil")
	}
	if scheme == nil {
		panic("scheme cannot be nil")
	}
	if metricsClient == nil {
		panic("metricsClient cannot be nil")
	}

	metricsWatcher := NewMetricsWatcher(metricsClient)
	if metricsWatcher == nil {
		panic("failed to create metrics watcher")
	}

	return &SelfRemediationPolicyReconciler{
		Client:         client,
		Scheme:         scheme,
		MetricsWatcher: metricsWatcher,
	}
}

// cleanupStaleRemediations removes tracking for resources that no longer exist
func (r *SelfRemediationPolicyReconciler) cleanupStaleRemediations(ctx context.Context) {
	log := log.FromContext(ctx)

	r.activeRemediations.Range(func(key, value interface{}) bool {
		state := value.(*RemediationState)

		// Check if policy still exists
		var policy remediationv1alpha1.SelfRemediationPolicy
		if err := r.Get(ctx, state.Policy, &policy); err != nil {
			if errors.IsNotFound(err) {
				log.Info("Cleaning up state for deleted policy",
					"policy", state.Policy.String())
				r.activeRemediations.Delete(key)
			}
			return true
		}

		// Check if target still exists
		var deployment appsv1.Deployment
		if err := r.Get(ctx, state.Target, &deployment); err != nil {
			if errors.IsNotFound(err) {
				log.Info("Cleaning up state for deleted target",
					"target", state.Target.String())
				r.activeRemediations.Delete(key)
			}
			return true
		}

		// Remove stale entries older than 1 hour
		if time.Since(state.LastChecked) > time.Hour {
			log.Info("Cleaning up stale remediation state",
				"policy", state.Policy.String(),
				"target", state.Target.String())
			r.activeRemediations.Delete(key)
		}

		return true
	})
}

// trackRemediation adds or updates tracking for an active remediation
func (r *SelfRemediationPolicyReconciler) trackRemediation(
	policy types.NamespacedName,
	target types.NamespacedName,
) {
	state := &RemediationState{
		LastChecked: time.Now(),
		Policy:      policy,
		Target:      target,
	}
	key := fmt.Sprintf("%s/%s", policy.String(), target.String())
	r.activeRemediations.Store(key, state)
}

// Reconcile handles the reconciliation loop for SelfRemediationPolicy
func (r *SelfRemediationPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Periodically cleanup stale remediations
	r.cleanupStaleRemediations(ctx)

	var policy remediationv1alpha1.SelfRemediationPolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get SelfRemediationPolicy")
		return ctrl.Result{}, err
	}

	// Get the pod referenced by the policy
	var pod corev1.Pod
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: policy.Spec.TargetRef.Namespace,
		Name:      policy.Spec.TargetRef.Name,
	}, &pod); err != nil {
		if errors.IsNotFound(err) {
			log.Info("Target pod not found", "pod", policy.Spec.TargetRef.Name)
			return ctrl.Result{RequeueAfter: time.Second * 30}, nil
		}
		log.Error(err, "unable to fetch target Pod")
		return ctrl.Result{}, err
	}

	// Parse CPU threshold
	threshold, err := strconv.ParseFloat(policy.Spec.CPUThreshold, 64)
	if err != nil {
		log.Error(err, "failed to parse CPU threshold")
		return ctrl.Result{}, err
	}

	// Check if pod is over threshold
	isOver, err := r.MetricsWatcher.IsPodOverThreshold(&pod, threshold)
	if err != nil {
		log.Error(err, "failed to check pod CPU threshold")
		return ctrl.Result{RequeueAfter: time.Second * 10}, nil
	}

	if isOver {
		// Process remediation rules
		for _, rule := range policy.Spec.Rules {
			if err := r.processRule(ctx, &policy, &pod, rule); err != nil {
				log.Error(err, "failed to process rule", "rule", rule.Name)
				continue
			}
		}
	}

	// Update status
	policy.Status.LastChecked = metav1.Now()
	policy.Status.Active = isOver
	if err := r.Status().Update(ctx, &policy); err != nil {
		log.Error(err, "failed to update policy status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Second * 30}, nil
}

func (r *SelfRemediationPolicyReconciler) processRule(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy, pod *corev1.Pod, rule remediationv1alpha1.Rule) error {

	for _, action := range rule.Actions {
		if action.Type == remediationv1alpha1.ScaleUp {
			var deployment appsv1.Deployment
			if err := r.Get(ctx, types.NamespacedName{
				Namespace: action.Target.Namespace,
				Name:      action.Target.Name,
			}, &deployment); err != nil {
				return fmt.Errorf("failed to get deployment: %w", err)
			}

			if err := r.executeActions(ctx, []remediationv1alpha1.Action{action}, &deployment); err != nil {
				return fmt.Errorf("failed to execute actions: %w", err)
			}

			// Track this remediation
			r.trackRemediation(types.NamespacedName{
				Namespace: policy.Namespace,
				Name:      policy.Name,
			}, types.NamespacedName{
				Namespace: deployment.Namespace,
				Name:      deployment.Name,
			})
		}
	}

	return nil
}

func (r *SelfRemediationPolicyReconciler) executeActions(
	ctx context.Context,
	actions []remediationv1alpha1.Action,
	deployment *appsv1.Deployment,
) error {
	log := log.FromContext(ctx)

	for _, action := range actions {
		actionLog := log.WithValues(
			"action_type", action.Type,
			"target_kind", action.Target.Kind,
			"target_name", action.Target.Name,
		)

		actionLog.Info("Executing remediation action")

		switch action.Type {
		case remediationv1alpha1.ScaleUp:
			// Skip if scaling parameters are not properly configured
			if action.ScalingParams == nil {
				actionLog.Info("Skipping action: scaling parameters not configured")
				continue
			}
			if action.ScalingParams.TemporaryMaxReplicas == nil {
				actionLog.Info("Skipping action: temporary max replicas not set")
				continue
			}

			// Store original replicas for later reversion
			originalReplicas := deployment.Spec.Replicas
			if originalReplicas == nil {
				// Set default if not specified
				defaultReplicas := int32(1)
				originalReplicas = &defaultReplicas
			}

			if deployment.Annotations == nil {
				deployment.Annotations = make(map[string]string)
			}
			deployment.Annotations["kubemedic.io/original-replicas"] = fmt.Sprintf("%d", *originalReplicas)

			// Scale up
			newReplicas := *action.ScalingParams.TemporaryMaxReplicas
			deployment.Spec.Replicas = &newReplicas

			actionLog.Info("Scaling up deployment",
				"original_replicas", *originalReplicas,
				"new_replicas", newReplicas,
				"scaling_duration", action.ScalingParams.ScalingDuration,
			)

			if err := r.Update(ctx, deployment); err != nil {
				actionLog.Error(err, "Failed to scale deployment")
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
