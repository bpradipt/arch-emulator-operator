/*


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

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	emulatorv1alpha1 "github.com/arch-emulator-operator/api/v1alpha1"
)

// ArchEmulatorReconciler reconciles a ArchEmulator object
type ArchEmulatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=emulator.multiarch.io,resources=archemulators,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=emulator.multiarch.io,resources=archemulators/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=emulator.multiarch.io,resources=archemulators/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch

func (r *ArchEmulatorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("archemulator", req.NamespacedName)

	// Fetch the ArchEmulator instance
	archemulator := &emulatorv1alpha1.ArchEmulator{}
	err := r.Get(ctx, req.NamespacedName, archemulator)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("ArchEmulator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get ArchEmulator")
		return ctrl.Result{}, err
	}
	// ArchEmulator instance found

	// Get nodeList based on selector
	nodeList := &corev1.NodeList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(archemulator.Spec.EmulatorNodeSelector.MatchLabels),
	}
	if err = r.List(ctx, nodeList, listOpts...); err != nil {
		log.Error(err, "Failed to list nodes", "archemulator.Name", archemulator.Name)
		return ctrl.Result{}, err
	}

	// Check if the Job already exists, if not create a new one
	found := &batchv1.Job{}
	err = r.Get(ctx, types.NamespacedName{Name: archemulator.Name, Namespace: archemulator.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Job
		job := r.jobForArchEmulator(archemulator, int32(len(nodeList.Items)))
		log.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name, "Job.Replicas", int32(len(nodeList.Items)))
		err = r.Create(ctx, job)
		if err != nil {
			log.Error(err, "Failed to create new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
			return ctrl.Result{}, err
		}
		// Job created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Job")
		return ctrl.Result{}, err
	}

	// Update the status

	// Set the emulator type
	archemulator.Status.EmulatorType = archemulator.Spec.EmulatorType

	archemulator.Status.Nodes = getNodeNames(nodeList.Items)

	err = r.Status().Update(ctx, archemulator)
	if err != nil {
		log.Error(err, "Failed to update archemulator status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// jobForArchEmulator returns an ArchEmulator Job object
func (r *ArchEmulatorReconciler) jobForArchEmulator(a *emulatorv1alpha1.ArchEmulator, numNodes int32) *batchv1.Job {

	ls := labelsForArchEmulator(a.Name)
	isPrivileged := true
	var ttl int32 = 10
	var backoffLimit int32 = 1

	var jobAffinity = corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{{
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{{
						Key:      "job-name",
						Operator: "In",
						Values:   []string{a.Name},
					}},
				},
				TopologyKey: "kubernetes.io/hostname",
			}},
		},
	}

	containerImage := "quay.io/bpradipt/qemu-user-static:latest"
	if a.Spec.EmulatorType.EmulatorImage != "" {
		containerImage = a.Spec.EmulatorType.EmulatorImage
	}
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.Name,
			Namespace: a.Namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism:             &numNodes,
			Completions:             &numNodes,
			TTLSecondsAfterFinished: &ttl,
			BackoffLimit:            &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           containerImage,
						Name:            a.Spec.EmulatorType.EmulatorName,
						Args:            []string{"--reset", "-p", "yes"},
						ImagePullPolicy: "IfNotPresent",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &isPrivileged,
						},
					}},
					NodeSelector:  a.Spec.EmulatorNodeSelector.MatchLabels,
					RestartPolicy: corev1.RestartPolicyNever,
					Affinity:      &jobAffinity,
				},
			},
		},
	}
	// Set ArchEmulator instance as the owner and controller
	ctrl.SetControllerReference(a, job, r.Scheme)
	return job
}

// Setup the resources watched by the controller
func (r *ArchEmulatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&emulatorv1alpha1.ArchEmulator{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}

// labelsForArchEmulator returns the labels for selecting the resources
// belonging to the given archemulator CR name.
func labelsForArchEmulator(name string) map[string]string {
	return map[string]string{"app": "archemulator", "archemulator_cr": name}
}

// Get NodeNames list
func getNodeNames(nodes []corev1.Node) []string {
	var nodeNames []string
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	return nodeNames
}
