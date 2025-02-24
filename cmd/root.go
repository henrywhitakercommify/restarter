package cmd

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/henrywhitakercommify/restarter/internal/k8s"
	"github.com/henrywhitakercommify/restarter/internal/log"
	"github.com/henrywhitakercommify/restarter/internal/metrics"
	"github.com/spf13/cobra"
)

var (
	kubeConfig   string
	namespace    string
	deployment   string
	restartWhen  float64
	restartAfter time.Duration
	interval     time.Duration
	dryRun       bool

	metricsPort int

	logLevel string
)

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restarter",
		Short: "Restart a kubernetes deployment when it is unhealthy",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Setup(log.Level(logLevel))
			server := metrics.Server(metricsPort)
			log.Info("serving metrics", "port", metricsPort)
			go server.ListenAndServe()
		},
		RunE: runRestarter,
	}

	cmd.Flags().
		StringVarP(&kubeConfig, "kube-config", "c", "~/.kube/config", "The path to your kubeconfig file")
	cmd.Flags().
		StringVarP(&namespace, "namespace", "n", "default", "The namespace the deployment is in")
	cmd.Flags().StringVarP(&deployment, "deployment", "d", "", "The name of the deployment")
	cmd.Flags().
		Float64VarP(&restartWhen, "restart-when", "k", 100, "The deployment will be rollout restarted when this percentage of pods are unready")
	cmd.Flags().
		DurationVarP(&interval, "interval", "i", time.Second*30, "The interval the ready status is evaluated at")
	cmd.Flags().
		BoolVar(&dryRun, "dry-run", false, "When enabled, the client will not restart the deployment, just log it")
	cmd.Flags().
		DurationVar(&restartAfter, "restart-after", time.Minute, "A time interval that the program waits for before checking again and restarting")

	cmd.PersistentFlags().
		IntVar(&metricsPort, "metrics-port", 8766, "The port the metrics server listens on")
	cmd.PersistentFlags().
		StringVar(&logLevel, "log-level", "info", "The log level, accepted values: info, error, debug")

	return cmd
}

func runRestarter(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient(kubeConfig)
	if err != nil {
		return err
	}

	dep := k8s.NewDeployment(client, namespace, deployment)
	if _, err := dep.Get(cmd.Context()); err != nil {
		return fmt.Errorf("deployment %s does not exist: %w", deployment, err)
	}

	tick := time.NewTicker(interval)
	defer tick.Stop()
	for {
		select {
		case <-cmd.Context().Done():
			return nil
		case <-tick.C:
			if restarting.Load() {
				log.Info("already checking deployment health, skipping")
				continue
			}
			go checkDeployment(cmd.Context(), dep)
		}
	}
}

var (
	restarting = &atomic.Bool{}
)

func checkDeployment(ctx context.Context, dep *k8s.Deployment) {
	restarting.Store(true)
	defer restarting.Store(false)

	labels := dep.Labels()

	metrics.TotalChecks.With(labels).Inc()
	ready, err := dep.Ready(ctx)
	if err != nil {
		log.Error("could not get ready status of deployment", "error", err)
		return
	}
	log.Info("got deployment ready status", "ready", fmt.Sprintf("%f%%", ready))
	if ready >= restartWhen {
		log.Info("deployment is healthy")
		return
	}

	metrics.TotalRestarts.With(labels).Inc()

	log.Info(
		"deployment ready status is less than threshold, waiting",
		"duration",
		restartAfter.String(),
	)

	time.Sleep(restartAfter)

	if dryRun {
		log.Info("deployment ready status is less than threshold but dry run is on, doing nothing")
		return
	}

	ready, err = dep.Ready(ctx)
	if err != nil {
		log.Error("could not get ready status of deployment", "error", err)
		return
	}

	if restartWhen < ready {
		log.Info("deployment is healthy again, skipping restart")
		return
	}

	if err := dep.Restart(ctx); err != nil {
		log.Error("failed to restart deployment", "error", err)
	}
}
