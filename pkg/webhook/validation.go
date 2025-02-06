package webhook

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	remediationv1alpha1 "kubemedic/api/v1alpha1"
)

// KubeMedicValidator handles validation of SelfRemediationPolicy resources
type KubeMedicValidator struct {
	Client  client.Client
	decoder admission.Decoder
}

// Handle validates SelfRemediationPolicy resources
func (v *KubeMedicValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	log := log.FromContext(ctx).WithValues(
		"webhook", "validator",
		"namespace", req.Namespace,
		"name", req.Name,
		"operation", req.Operation,
	)

	log.Info("Handling admission request")

	policy := &remediationv1alpha1.SelfRemediationPolicy{}
	err := v.decoder.Decode(req, policy)
	if err != nil {
		log.Error(err, "Failed to decode admission request")
		return admission.Errored(http.StatusBadRequest, err)
	}

	if err := v.validatePolicy(ctx, policy); err != nil {
		log.Error(err, "Policy validation failed")
		return admission.Denied(err.Error())
	}

	log.Info("Policy validation successful")
	return admission.Allowed("")
}

// InjectDecoder injects the decoder
func (v *KubeMedicValidator) InjectDecoder(d *admission.Decoder) error {
	if d == nil {
		return fmt.Errorf("decoder cannot be nil")
	}
	v.decoder = *d
	return nil
}

// validatePolicy validates all aspects of the policy
func (v *KubeMedicValidator) validatePolicy(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	log := log.FromContext(ctx).WithValues(
		"policy_name", policy.Name,
		"rules_count", len(policy.Spec.Rules),
	)

	log.V(1).Info("Starting policy validation")

	if err := v.validateNamespace(ctx, policy); err != nil {
		log.Error(err, "Namespace validation failed")
		return err
	}

	if err := v.validateResources(ctx, policy); err != nil {
		log.Error(err, "Resource validation failed")
		return err
	}

	if err := v.validateActions(ctx, policy); err != nil {
		log.Error(err, "Action validation failed")
		return err
	}

	if err := v.validateSafetyLimits(ctx, policy); err != nil {
		log.Error(err, "Safety limits validation failed")
		return err
	}

	log.Info("Policy validation completed successfully")
	return nil
}

func (v *KubeMedicValidator) validateNamespace(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
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

func (v *KubeMedicValidator) validateResources(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	allowedResources := map[string]bool{
		"deployments":              true,
		"statefulsets":             true,
		"horizontalpodautoscalers": true,
	}

	for _, rule := range policy.Spec.Rules {
		for _, action := range rule.Actions {
			// Check if Target is initialized
			if action.Target.Kind == "" || action.Target.Name == "" {
				return fmt.Errorf("action target must specify both kind and name")
			}

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

func (v *KubeMedicValidator) validateActions(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
	log := log.FromContext(ctx)

	for _, rule := range policy.Spec.Rules {
		ruleLog := log.WithValues(
			"rule_name", rule.Name,
			"actions_count", len(rule.Actions),
		)

		for _, action := range rule.Actions {
			actionLog := ruleLog.WithValues(
				"action_type", action.Type,
				"target_kind", action.Target.Kind,
				"target_name", action.Target.Name,
			)

			actionLog.V(1).Info("Validating action")

			if err := v.validateActionParams(action); err != nil {
				actionLog.Error(err, "Action parameters validation failed")
				return err
			}

			if err := v.validateTargetResource(ctx, action.Target); err != nil {
				actionLog.Error(err, "Target resource validation failed")
				return err
			}

			// For scaling actions, validate quota limits
			if action.Type == remediationv1alpha1.ScaleUp {
				if err := v.validateQuotas(ctx, policy); err != nil {
					actionLog.Error(err, "Quota validation failed")
					return err
				}
			}

			actionLog.V(1).Info("Action validation successful")
		}
	}

	return nil
}

func (v *KubeMedicValidator) validateSafetyLimits(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
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

func (v *KubeMedicValidator) validateQuotas(ctx context.Context, policy *remediationv1alpha1.SelfRemediationPolicy) error {
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

func (v *KubeMedicValidator) validateTargetResource(ctx context.Context, target remediationv1alpha1.Target) error {
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

func (v *KubeMedicValidator) validateActionParams(action remediationv1alpha1.Action) error {
	switch action.Type {
	case remediationv1alpha1.ScaleUp, remediationv1alpha1.ScaleDown:
		if action.ScalingParams == nil {
			return fmt.Errorf("scaling parameters required for action type %s", action.Type)
		}
		if action.ScalingParams.TemporaryMaxReplicas == nil {
			return fmt.Errorf("temporaryMaxReplicas must be specified for scaling actions")
		}
	case remediationv1alpha1.AdjustHPALimits:
		if action.ScalingParams == nil || action.ScalingParams.TemporaryMaxReplicas == nil {
			return fmt.Errorf("temporary max replicas required for HPA adjustment")
		}
	}

	// Validate scaling duration if specified
	if action.ScalingParams != nil && action.ScalingParams.ScalingDuration != "" {
		if _, err := time.ParseDuration(action.ScalingParams.ScalingDuration); err != nil {
			return fmt.Errorf("invalid scaling duration format: %v", err)
		}
	}

	return nil
}

func (v *KubeMedicValidator) getCurrentReplicas(ctx context.Context, target remediationv1alpha1.Target) (int32, error) {
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

func (v *KubeMedicValidator) checkQuotaLimits(ctx context.Context, action remediationv1alpha1.Action, quota *corev1.ResourceQuota) error {
	requirements, err := v.getResourceRequirements(ctx, action.Target)
	if err != nil {
		return err
	}

	// Calculate new resource usage
	for resourceName, proposed := range requirements {
		if hard, exists := quota.Status.Hard[resourceName]; exists {
			// Convert to int64 for comparison
			proposedValue := proposed.Value()
			hardValue := hard.Value()

			if proposedValue > hardValue {
				return fmt.Errorf("action would exceed quota for %s", resourceName)
			}
		}
	}

	return nil
}

func (v *KubeMedicValidator) getResourceRequirements(ctx context.Context, target remediationv1alpha1.Target) (corev1.ResourceList, error) {
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
