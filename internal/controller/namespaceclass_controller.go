package controller

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/resource-wrapper/api/v1alpha1"
	"golang.org/x/time/rate"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +kubebuilder:rbac:groups=core.resource-wrapper.io,resources=namespaceclasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.resource-wrapper.io,resources=namespaceclasses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.resource-wrapper.io,resources=namespaceclasses/finalizers,verbs=update
// +kubebuilder:rbac:groups=core,resources=namespaces,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete

// NamespaceClassReconciler reconciles a NamespaceClass object
type NamespaceClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Create a new rate limiter that combines exponential backoff with a maximum rate
	rateLimiter := workqueue.NewTypedMaxOfRateLimiter[reconcile.Request](
		workqueue.NewTypedItemExponentialFailureRateLimiter[reconcile.Request](5*time.Millisecond, 30*time.Second),
		&workqueue.TypedBucketRateLimiter[reconcile.Request]{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)

	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			RateLimiter:             rateLimiter,
			MaxConcurrentReconciles: 5,
		}).
		For(&v1alpha1.NamespaceClass{}).
		Watches(
			&corev1.Namespace{},
			handler.EnqueueRequestsFromMapFunc(r.findNamespaceClassForNamespace),
		).
		Watches(
			&corev1.ServiceAccount{},
			handler.EnqueueRequestsFromMapFunc(r.findNamespaceClassForNamespace),
		).
		Watches(
			&networkingv1.NetworkPolicy{},
			handler.EnqueueRequestsFromMapFunc(r.findNamespaceClassForNamespace),
		).
		WithEventFilter(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			// Skip the filter for NamespaceClass objects
			if _, ok := obj.(*v1alpha1.NamespaceClass); ok {
				return true
			}
			// For other objects, only process if they have our label
			return obj.GetLabels()["namespaceclass.akuity.io/name"] != ""
		})).
		Complete(r)
}

// findNamespaceClassForNamespace maps a Namespace to its NamespaceClass
func (r *NamespaceClassReconciler) findNamespaceClassForNamespace(ctx context.Context, obj client.Object) []reconcile.Request {
	// Get the namespace class name from the annotation
	className := obj.GetLabels()["namespaceclass.akuity.io/name"]
	if className == "" {
		return nil
	}

	return []reconcile.Request{
		{NamespacedName: types.NamespacedName{Name: className}},
	}
}

func (r *NamespaceClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling NamespaceClass", "name", req.Name)

	instance := &v1alpha1.NamespaceClass{}
	logger.Info("Fetching NamespaceClass", "name", req.Name)

	err := r.Client.Get(ctx, client.ObjectKey{Name: req.NamespacedName.Name}, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Error(err, "NamespaceClass not Found")
			return ctrl.Result{}, nil
		}

		logger.Error(err, "Failed to get NamespaceClass")
		return ctrl.Result{}, err
	}

	logger = logger.WithValues("NamespaceClass", instance.Name)

	// List all namespaces with this class
	namespaces := &corev1.NamespaceList{}

	if err := r.List(ctx, namespaces, client.MatchingLabels{
		"namespaceclass.akuity.io/name": instance.Name,
	}); err != nil {
		logger.Error(err, "Failed to list namespaces")
		return ctrl.Result{}, err
	}

	// Process each namespace
	for _, ns := range namespaces.Items {
		logger := logger.WithValues("namespace", ns.Name)

		// Handle NetworkPolicy
		if instance.Spec.NetworkPolicyTemplate != nil {
			if err := r.reconcileNetworkPolicy(ctx, instance, &ns); err != nil {
				logger.Error(err, "Failed to reconcile NetworkPolicy")
				return ctrl.Result{}, err
			}
		}

		// Handle ServiceAccount
		if instance.Spec.ServiceAccountTemplate != nil {
			if err := r.reconcileServiceAccount(ctx, instance, &ns); err != nil {
				logger.Error(err, "Failed to reconcile ServiceAccount")
				return ctrl.Result{}, err
			}
		}
	}

	// Update status
	instance.Status.ObservedGeneration = instance.Generation
	instance.Status.LastAppliedTime = v1.Now()
	if err := r.Client.Status().Patch(ctx, instance, client.MergeFrom(instance.DeepCopy())); err != nil {
		if apierrors.IsNotFound(err) {
			// The resource was deleted, which is fine
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to patch status")
		return ctrl.Result{}, err
	}

	logger.Info("NamespaceClass reconciled")
	return ctrl.Result{}, nil
}

func (r *NamespaceClassReconciler) reconcileNetworkPolicy(ctx context.Context, class *v1alpha1.NamespaceClass, ns *corev1.Namespace) error {
	desired := &networkingv1.NetworkPolicy{
		ObjectMeta: v1.ObjectMeta{
			Name:      "admin-network-policy",
			Namespace: ns.Name,
			Labels: map[string]string{
				"namespaceclass.akuity.io/name": class.Name,
			},
		},
		Spec: *class.Spec.NetworkPolicyTemplate.DeepCopy(),
	}

	// Create or update the NetworkPolicy
	existing := &networkingv1.NetworkPolicy{}
	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: ns.Name}, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, desired)
		}
		return err
	}

	if existing.GetObjectMeta().GetLabels()["namespaceclass.akuity.io/name"] != class.Name {
		existing.GetObjectMeta().SetLabels(map[string]string{
			"namespaceclass.akuity.io/name": class.Name,
		})
	}

	// Update existing NetworkPolicy
	existing.Spec = desired.Spec
	return r.Update(ctx, existing)
}

func (r *NamespaceClassReconciler) reconcileServiceAccount(ctx context.Context, class *v1alpha1.NamespaceClass, ns *corev1.Namespace) error {
	desired := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:      "admin-service-account",
			Namespace: ns.Name,
			Labels: map[string]string{
				"namespaceclass.akuity.io/name": class.Name,
			},
		},
		AutomountServiceAccountToken: class.Spec.ServiceAccountTemplate.AutomountServiceAccountToken,
	}

	// Create or update the ServiceAccount
	existing := &corev1.ServiceAccount{}
	err := r.Get(ctx, types.NamespacedName{Name: desired.Name, Namespace: ns.Name}, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return r.Create(ctx, desired)
		}
		return err
	}

	if existing.GetObjectMeta().GetLabels()["namespaceclass.akuity.io/name"] != class.Name {
		existing.GetObjectMeta().SetLabels(map[string]string{
			"namespaceclass.akuity.io/name": class.Name,
		})
	}

	// Update existing ServiceAccount
	existing.AutomountServiceAccountToken = desired.AutomountServiceAccountToken
	return r.Update(ctx, existing)
}
