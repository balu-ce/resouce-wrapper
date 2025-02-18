package main

import (
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/resource-wrapper/api/v1alpha1"
	"github.com/resource-wrapper/internal/controller"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

func main() {
	var (
		app   = kingpin.New("resource-wrapper", "resource-wrapper").DefaultEnvars()
		debug = app.Flag("debug", "Enable debug mode").Default("true").Bool()
	)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	ctrl.SetLogger(zap.New(zap.UseDevMode(*debug)))

	cfg, err := ctrl.GetConfig()
	if err != nil {
		kingpin.FatalIfError(err, "Cannot get config")
	}

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		kingpin.FatalIfError(err, "Cannot create controller manager")
	}

	if err := (&controller.NamespaceClassReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    ctrl.Log.WithName("controllers").WithName("NamespaceClass"),
	}).SetupWithManager(mgr); err != nil {
		kingpin.FatalIfError(err, "Cannot create controller")
	}

	kingpin.FatalIfError(mgr.Start(ctrl.SetupSignalHandler()), "Cannot start controller manager")
}
