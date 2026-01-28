package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/ikepcampbell/kubemedic/internal/version"
	webhookpkg "github.com/ikepcampbell/kubemedic/pkg/webhook"
)

func main() {
	var tlsCertFile string
	var tlsKeyFile string
	var printVersion bool

	flag.StringVar(&tlsCertFile, "tls-cert-file", "/etc/webhook/certs/tls.crt", "File containing the x509 Certificate for HTTPS")
	flag.StringVar(&tlsKeyFile, "tls-private-key-file", "/etc/webhook/certs/tls.key", "File containing the x509 private key")
	flag.BoolVar(&printVersion, "version", false, "Print version information and exit")
	flag.Parse()

	if printVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	certDir := filepath.Dir(tlsCertFile)
	keyDir := filepath.Dir(tlsKeyFile)
	if certDir != keyDir {
		klog.ErrorS(nil, "tls-cert-file and tls-private-key-file must be in the same directory", "tls-cert-file", tlsCertFile, "tls-private-key-file", tlsKeyFile)
		os.Exit(1)
	}
	certName := filepath.Base(tlsCertFile)
	keyName := filepath.Base(tlsKeyFile)

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		klog.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}

	certDir := filepath.Dir(tlsCertFile)
	certName := filepath.Base(tlsCertFile)
	keyDir := filepath.Dir(tlsKeyFile)
	keyName := filepath.Base(tlsKeyFile)
	if certDir != keyDir {
		klog.Error(fmt.Errorf("tls cert and key must be in the same directory"), "invalid TLS flags", "tls-cert-file", tlsCertFile, "tls-private-key-file", tlsKeyFile)
		os.Exit(1)
	}

	// Create a new manager to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		WebhookServer: webhook.NewServer(webhook.Options{
			Port:     8443,
			CertDir:  certDir,
			CertName: certName,
			KeyName:  keyName,
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

	klog.InfoS("starting webhook server", "version", version.String())
	if err := mgr.Start(ctx); err != nil {
		klog.Error(err, "error starting webhook server")
		os.Exit(1)
	}
}
