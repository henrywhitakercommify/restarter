package cmd

import (
	"fmt"
	"time"

	"github.com/henrywhitakercommify/restarter/internal/k8s"
	"github.com/henrywhitakercommify/restarter/internal/log"
	"github.com/spf13/cobra"
)

var (
	kubeConfig  string
	namespace   string
	deployment  string
	restartWhen float64
	interval    time.Duration
	dryRun      bool

	logLevel string
)

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restarter",
		Short: "Restart a kubernetes deployment when it is unhealthy",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.Setup(log.Level(logLevel))
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
		DurationVarP(&interval, "interval", "i", time.Second*5, "The interval the ready status is evaluated at")
	cmd.Flags().
		BoolVar(&dryRun, "dry-run", false, "When enabled, the client will not restart the deployment, just log it")
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
			ready, err := dep.Ready(cmd.Context())
			if err != nil {
				log.Error("could not get ready status of deployment", "error", err)
			}
			log.Info("got deployment ready status", "ready", fmt.Sprintf("%f%%", ready))
			if !dryRun && ready < restartWhen {
				log.Info("deployment ready status is less than threshold, resting deployment")
				if err := dep.Restart(cmd.Context()); err != nil {
					log.Error("failed to restart deployment", "error", err)
					continue
				}
			}
		}
	}
}
