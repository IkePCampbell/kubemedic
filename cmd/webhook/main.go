package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	webhookpkg "kubemedic/pkg/webhook"
)

func main() {
	var tlsCertFile string
	var tlsKeyFile string

	flag.StringVar(&tlsCertFile, "tls-cert-file", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS")
	flag.StringVar(&tlsKeyFile, "tls-private-key-file", "/etc/webhook/certs/tls.key", "File containing the x509 private key")
	flag.Parse()

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}

	// Create a new manager to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:    8443,
			CertDir: "/etc/webhook/certs",
		}),
		Metrics: metricsserver.Options{
			BindAddress: "0",
		},
	})
	if err != nil {
		klog.Error(err, "unable to create manager")
		os.Exit(1)
	}

	// Create and initialize the validator
	validator := &webhookpkg.KubeMedicValidator{
		Client: mgr.GetClient(),
	}

	// Register the webhook with the manager
	mgr.GetWebhookServer().Register("/validate", &admission.Webhook{
		Handler: validator,
	})

	// Create a context that is canceled when a termination signal is received
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	klog.Info("starting webhook server")
	if err := mgr.Start(ctx); err != nil {
		klog.Error(err, "error starting webhook server")
		os.Exit(1)
	}
}
