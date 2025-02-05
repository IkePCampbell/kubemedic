package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	remediationv1alpha1 "github.com/ikepcampbell/kubemedic/api/v1alpha1"
)

// PolicyValidator handles validation of SelfRemediationPolicy resources
type PolicyValidator struct {
	Client  client.Client
	decoder *admission.Decoder
}

// Handle implements admission.Handler
func (v *PolicyValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	policy := &remediationv1alpha1.SelfRemediationPolicy{}

	err := v.decoder.Decode(req, policy)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Run all validations
	if err := v.validatePolicy(ctx, policy); err != nil {
		return admission.Denied(err.Error())
	}

	return admission.Allowed("")
}

// validatePolicy runs all validation checks
func (v *PolicyValidator) validatePolicy(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	// Namespace restrictions
	if err := v.validateNamespace(ctx, policy); err != nil {
		return err
	}

	// Resource restrictions
	if err := v.validateResources(ctx, policy); err != nil {
		return err
	}

	// Action restrictions
	if err := v.validateActions(ctx, policy); err != nil {
		return err
	}

	// Safety limits
	if err := v.validateSafetyLimits(ctx, policy); err != nil {
		return err
	}

	// Quota validation
	if err := v.validateQuotas(ctx, policy); err != nil {
		return err
	}

	return nil
}

func (v *PolicyValidator) validateNamespace(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	// Check if namespace is in denied list
	deniedNamespaces := []string{
		"kube-system",
		"kube-public",
		"kube-node-lease",
		"cert-manager",
		"ingress-nginx",
	}

	for _, ns := range deniedNamespaces {
		if policy.Namespace == ns {
			return fmt.Errorf("namespace %s is not allowed for remediation policies", ns)
		}
	}

	// Check if namespace has required labels/annotations
	namespace := &corev1.Namespace{}
	if err := v.Client.Get(ctx, client.ObjectKey{Name: policy.Namespace}, namespace); err != nil {
		return err
	}

	if namespace.Labels["kubemedic.io/exclude"] == "true" {
		return fmt.Errorf("namespace is excluded from remediation")
	}

	return nil
}

func (v *PolicyValidator) validateResources(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	allowedResources := map[string]bool{
		"deployments":              true,
		"statefulsets":            true,
		"horizontalpodautoscalers": true,
	}

	for _, rule := range policy.Spec.Rules {
		for _, action := range rule.Actions {
			if !allowedResources[strings.ToLower(action.Target.Kind)] {
				return fmt.Errorf("resource type %s is not allowed", action.Target.Kind)
			}

			// Check if target resource exists and is not protected
			if err := v.validateTargetResource(ctx, action.Target); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *PolicyValidator) validateActions(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	allowedActions := map[string]bool{
		"ScaleUp":          true,
		"ScaleDown":        true,
		"RestartPod":       true,
		"AdjustHPALimits": true,
	}

	for _, rule := range policy.Spec.Rules {
		for _, action := range rule.Actions {
			if !allowedActions[action.Type] {
				return fmt.Errorf("action type %s is not allowed", action.Type)
			}

			// Validate action-specific parameters
			if err := v.validateActionParams(action); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *PolicyValidator) validateSafetyLimits(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	// Global limits
	maxScaleFactor := 2
	maxDuration := 2 * time.Hour
	minPods := 1

	for _, rule := range policy.Spec.Rules {
		for _, action := range rule.Actions {
			if action.ScalingParams != nil {
				// Check scale factor
				if action.ScalingParams.TemporaryMaxReplicas != nil {
					currentReplicas, err := v.getCurrentReplicas(ctx, action.Target)
					if err != nil {
						return err
					}

					if *action.ScalingParams.TemporaryMaxReplicas > currentReplicas*int32(maxScaleFactor) {
						return fmt.Errorf("scale factor exceeds maximum allowed (%d)", maxScaleFactor)
					}

					if *action.ScalingParams.TemporaryMaxReplicas < int32(minPods) {
						return fmt.Errorf("minimum pods cannot be less than %d", minPods)
					}
				}

				// Check duration
				if action.ScalingParams.ScalingDuration != "" {
					duration, err := time.ParseDuration(action.ScalingParams.ScalingDuration)
					if err != nil {
						return fmt.Errorf("invalid duration format: %v", err)
					}

					if duration > maxDuration {
						return fmt.Errorf("scaling duration exceeds maximum allowed (%v)", maxDuration)
					}
				}
			}
		}
	}

	return nil
}

func (v *PolicyValidator) validateQuotas(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	// Get namespace quotas
	resourceQuota := &corev1.ResourceQuota{}
	if err := v.Client.Get(ctx, client.ObjectKey{
		Name:      "compute-resources",
		Namespace: policy.Namespace,
	}, resourceQuota); err != nil {
		// If no quota exists, skip validation
		return nil
	}

	// Calculate potential resource usage
	for _, rule := range policy.Spec.Rules {
		for _, action := range rule.Actions {
			if action.ScalingParams != nil && action.ScalingParams.TemporaryMaxReplicas != nil {
				// Check if scaling would exceed quota
				if err := v.checkQuotaLimits(ctx, action, resourceQuota); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (v *PolicyValidator) validateTargetResource(ctx context.Context, target remediationv1alpha1.Target) error {
	// Get the target resource
	obj := &unstructured.Unstructured{}
	obj.SetKind(target.Kind)
	obj.SetNamespace(target.Namespace)
	obj.SetName(target.Name)

	if err := v.Client.Get(ctx, client.ObjectKey{
		Name:      target.Name,
		Namespace: target.Namespace,
	}, obj); err != nil {
		return fmt.Errorf("target resource not found: %v", err)
	}

	// Check protection labels
	if obj.GetLabels()["kubemedic.io/protected"] == "true" {
		return fmt.Errorf("target resource is protected from remediation")
	}

	return nil
}

func (v *PolicyValidator) validateActionParams(action remediationv1alpha1.Action) error {
	switch action.Type {
	case "ScaleUp", "ScaleDown":
		if action.ScalingParams == nil {
			return fmt.Errorf("scaling parameters required for action type %s", action.Type)
		}
	case "AdjustHPALimits":
		if action.ScalingParams == nil || action.ScalingParams.TemporaryMaxReplicas == nil {
			return fmt.Errorf("temporary max replicas required for HPA adjustment")
		}
	}

	return nil
}

func (v *PolicyValidator) getCurrentReplicas(ctx context.Context, target remediationv1alpha1.Target) (int32, error) {
	switch target.Kind {
	case "Deployment":
		deploy := &appsv1.Deployment{}
		if err := v.Client.Get(ctx, client.ObjectKey{
			Name:      target.Name,
			Namespace: target.Namespace,
		}, deploy); err != nil {
			return 0, err
		}
		return *deploy.Spec.Replicas, nil

	case "StatefulSet":
		sts := &appsv1.StatefulSet{}
		if err := v.Client.Get(ctx, client.ObjectKey{
			Name:      target.Name,
			Namespace: target.Namespace,
		}, sts); err != nil {
			return 0, err
		}
		return *sts.Spec.Replicas, nil

	default:
		return 0, fmt.Errorf("unsupported resource type for replica count: %s", target.Kind)
	}
}

func (v *PolicyValidator) checkQuotaLimits(ctx context.Context, action remediationv1alpha1.Action, quota *corev1.ResourceQuota) error {
	// Get resource requirements for the target
	requirements, err := v.getResourceRequirements(ctx, action.Target)
	if err != nil {
		return err
	}

	// Calculate total resource usage after scaling
	proposedReplicas := *action.ScalingParams.TemporaryMaxReplicas
	for resourceName, quantity := range requirements {
		proposed := quantity.DeepCopy()
		proposed.MultiplyInt64(int64(proposedReplicas))

		if quota.Status.Hard[resourceName] != nil {
			if proposed.Cmp(*quota.Status.Hard[resourceName]) > 0 {
				return fmt.Errorf("scaling would exceed quota for %s", resourceName)
			}
		}
	}

	return nil
}

func (v *PolicyValidator) getResourceRequirements(ctx context.Context, target remediationv1alpha1.Target) (corev1.ResourceList, error) {
	switch target.Kind {
	case "Deployment":
		deploy := &appsv1.Deployment{}
		if err := v.Client.Get(ctx, client.ObjectKey{
			Name:      target.Name,
			Namespace: target.Namespace,
		}, deploy); err != nil {
			return nil, err
		}
		return calculatePodResources(deploy.Spec.Template.Spec.Containers), nil

	case "StatefulSet":
		sts := &appsv1.StatefulSet{}
		if err := v.Client.Get(ctx, client.ObjectKey{
			Name:      target.Name,
			Namespace: target.Namespace,
		}, sts); err != nil {
			return nil, err
		}
		return calculatePodResources(sts.Spec.Template.Spec.Containers), nil

	default:
		return nil, fmt.Errorf("unsupported resource type for requirements: %s", target.Kind)
	}
}

func calculatePodResources(containers []corev1.Container) corev1.ResourceList {
	result := corev1.ResourceList{}
	for _, container := range containers {
		for resourceName, quantity := range container.Resources.Requests {
			if current, exists := result[resourceName]; exists {
				current.Add(quantity)
				result[resourceName] = current
			} else {
				result[resourceName] = quantity.DeepCopy()
			}
		}
	}
	return result
}

// InjectDecoder implements admission.DecoderInjector
func (v *PolicyValidator) InjectDecoder(d *admission.Decoder) error {
	v.decoder = d
	return nil
} 