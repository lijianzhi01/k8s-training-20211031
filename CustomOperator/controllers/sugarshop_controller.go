/*
Copyright 2021.

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
	"strconv"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webappv1 "julian.swat/sugarshop/api/v1"
)

// SugarshopReconciler reconciles a Sugarshop object
type SugarshopReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.julian.swat,resources=sugarshops,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.julian.swat,resources=sugarshops/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.julian.swat,resources=sugarshops/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Sugarshop object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *SugarshopReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// your logic here
	snapshot := &webappv1.Sugarshop{}
	snapshotReadErr := r.Get(ctx, req.NamespacedName, snapshot)

	// Exit early if something went wrong parsing the SugarSnapshot
	if snapshotReadErr != nil || len(snapshot.Name) == 0 {
		r.Log.Error(snapshotReadErr, "Error encountered reading SugarSnapshot")
		return ctrl.Result{}, snapshotReadErr
	}

	snapshotCreator := &corev1.Pod{}
	snapshotCreatorReadErr := r.Get(
		ctx, types.NamespacedName{Name: "snapshot-creator", Namespace: req.Namespace}, snapshotCreator,
	)

	// Exit early if the SugarSnapshot creator is already running
	if snapshotCreatorReadErr == nil {
		r.Log.Error(snapshotCreatorReadErr, "Snapshot creator already running!")
		return ctrl.Result{}, snapshotCreatorReadErr
	}

	newPvName := "sugar-snapshot-pv-" + strconv.FormatInt(time.Now().Unix(), 10) + "0" + snapshot.Spec.SourceVolumeName
	newPvcName := "sugar-snapshot-pvc-" + strconv.FormatInt(time.Now().Unix(), 10) + "0" + snapshot.Spec.SourceClaimName

	// Create a new Persistent Volume
	newPersistentVol := corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newPvName,
			Namespace: req.Namespace,
		},
		Spec: corev1.PersistentVolumeSpec{
			StorageClassName: "manual",
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Capacity: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("1Gi"),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: snapshot.Spec.HostPath,
				},
			},
		},
	}

	pvCreateErr := r.Create(ctx, &newPersistentVol)
	if pvCreateErr != nil {
		r.Log.Error(pvCreateErr, "Error encountered creating new pvc")
		return ctrl.Result{}, pvCreateErr
	}

	_ = r.Log.WithValues("Created New Snapshot Persistent Volume", req.NamespacedName)

	manualStorageClass := "manual"

	// Create a new Persistent Volume Claim connected to the new Persistent Volume
	newPersistentVolClaim := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      newPvcName,
			Namespace: req.Namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &manualStorageClass,
			VolumeName:       newPvName,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceName(corev1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
		},
	}

	pvcCreateErr := r.Create(ctx, &newPersistentVolClaim)
	if pvcCreateErr != nil {
		r.Log.Error(pvcCreateErr, "Error encountered creating new pvc")
		return ctrl.Result{}, pvcCreateErr
	}

	_ = r.Log.WithValues("Created New Snapshot Persistent Volume Claim", req.NamespacedName)

	// Start a Pod that is hooked up to the snapshot PVC and the original PVC
	// The Pod simply copies a file from the old PVC to the new PVC - creating a primitive snapshot
	snapshotCreatorPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snapshot-creator",
			Namespace: req.Namespace,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: "Never",
			Volumes: []corev1.Volume{
				{
					Name: "sugar-snapshot-" + snapshot.Spec.SourceClaimName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: snapshot.Spec.SourceClaimName,
						},
					},
				},
				{
					Name: "sugar-snapshot-" + newPvcName,
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: newPvcName,
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:  "busybox",
					Image: "ubuntu",
					Command: []string{
						"/bin/sh",
						"-c",
						"cp /tmp/source/test.file /tmp/dest/test.file",
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "sugar-snapshot-" + snapshot.Spec.SourceClaimName,
							MountPath: "/tmp/source",
						},
						{
							Name:      "sugar-snapshot-" + newPvcName,
							MountPath: "/tmp/dest",
						},
					},
				},
			},
		},
	}

	podCreateErr := r.Create(ctx, &snapshotCreatorPod)
	if podCreateErr != nil {
		r.Log.Error(podCreateErr, "Error encountered creating snapshotting pod")
		return ctrl.Result{}, podCreateErr
	}

	_ = r.Log.WithValues("Instantiating snapshot-creator pod", req.NamespacedName)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SugarshopReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Sugarshop{}).
		Complete(r)
}
