package controllers

import (
	"context"
	"time"

	api "github.com/keikoproj/iam-manager/api/v1alpha1"
	"github.com/keikoproj/iam-manager/internal/config"
	"github.com/keikoproj/iam-manager/pkg/logging"
	ctrl "sigs.k8s.io/controller-runtime"
)

/**
 * This will start a go routine that will run every "ControllerDesiredFrequency" seconds.
 */
func (r *IamroleReconciler) StartControllerReconcileCronJob(ctx context.Context) error {
	log := logging.Logger(ctx, "controllers", "iamrole_controller", "StartControllerReconcileCronJob")

	log.Info("Starting the cronjob")

	// If the controllerDesiredFrequency is less than the minimum, use the minimum.
	// This is to prevent the controller from running too frequently.
	// DesiredFrequency can be configured in iks-config
	controllerDesiredFrequency := config.ControllerMinimumDesiredFrequency
	if controllerDesiredFrequency < config.Props.ControllerDesiredFrequency() {
		controllerDesiredFrequency = config.Props.ControllerDesiredFrequency()
	}

	log.Info("StartControllerReconcileCronJob", "controllerDesiredFrequency", controllerDesiredFrequency)

	ticker := time.NewTicker(time.Duration(controllerDesiredFrequency) * time.Second)

	for {
		select {
		case <-ticker.C:
			start := time.Now()
			log.Info("StartControllerReconcileCronJob - start to fetch IAM roles", "time", time.Now())

			r.ReconcileAllReadyStateIamRoles(ctx)

			log.Info("Reconciled all iam-roles", "time", time.Now(), "duration", time.Since(start))
		case <-ctx.Done():
			log.Info("Application graceful shutdown", "time", time.Now())
			return nil
		}
	}
}

func (r *IamroleReconciler) ReconcileAllReadyStateIamRoles(ctx context.Context) {
	log := logging.Logger(ctx, "controllers", "iamrole_controller", "Worker")

	var err error
	var res ctrl.Result
	var iamRoles []*api.Iamrole
	var iamrole *api.Iamrole

	if iamRoles, err = api.ListIamRoles(context.Background(), r.Client); err != nil {
		log.Error(err, "StartControllerReconcileCronJob", "unable to list iamroles CR")
		return
	}

	for _, prefetchedIamRole := range iamRoles {
		log.Info("Reconcile start", "iamRole", prefetchedIamRole.Name)
		if iamrole, err = api.GetIamRole(context.Background(), r.Client, prefetchedIamRole.Name, prefetchedIamRole.Namespace); err != nil {
			log.Error(err, "unable to get iamrole resource", "iamRole", prefetchedIamRole.Name, "namespace", prefetchedIamRole.Namespace)
			continue
		}

		if iamrole.Status.State != "Ready" {
			log.Info("Reconcile skipped because its state is not ready", "iamRole", iamrole.Name, "state", iamrole.Status.State)
			continue
		}

		res, err = r.HandleReconcile(ctx, ctrl.Request{}, iamrole)
		log.Info("Reconcile result", "result", res, "error", err)

		// sleep for 2 seconds for politeness
		time.Sleep(2 * time.Second)
	}
}
