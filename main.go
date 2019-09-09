/*
Copyright (C) 2019 Synopsys, Inc.

Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"flag"
	"fmt"
	"github.com/blackducksoftware/synopsys-operator/utils"
	"github.com/spf13/viper"
	"os"
	"strings"

	synopsysv1 "github.com/blackducksoftware/synopsys-operator/api/v1"
	"github.com/blackducksoftware/synopsys-operator/controllers"
	controllers_utils "github.com/blackducksoftware/synopsys-operator/controllers/util"
	routev1 "github.com/openshift/api/route/v1"
	securityv1 "github.com/openshift/api/security/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = synopsysv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

type config struct {
	LogLevel        string
	Namespace       string
	CrdNames        string
	IsClusterScoped bool
}

func main() {
	var operatorConfig *config

	if len(os.Args) == 0 {
		panic("config file is missing")
	}

	// Get config
	configPath := os.Args[1]
	viper.SetConfigFile(configPath)
	err := viper.ReadInConfig()
	if err != nil {

		return
	}
	err = viper.Unmarshal(&operatorConfig)
	if err != nil {
		setupLog.Error(err, "failed to unmarshal config")
		return
	}

	var crdNamespace string
	if !operatorConfig.IsClusterScoped {
		crdNamespace = operatorConfig.Namespace
	}

	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Namespace:          crdNamespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	isOpenShift := controllers_utils.IsOpenShift(mgr.GetConfig())
	setupLog.V(1).Info("cluster configuration", "isOpenShift", isOpenShift)

	if isOpenShift {
		_ = routev1.AddToScheme(scheme)
		_ = securityv1.AddToScheme(scheme)
	}

	// Check which controllers are enabled
	if len(operatorConfig.CrdNames) > 0 {
		crds := strings.Split(operatorConfig.CrdNames, ",")
		for _, crd := range crds {
			switch strings.TrimSpace(crd) {
			case utils.AlertCRDName:
				if err = (&controllers.AlertReconciler{
					Client:      mgr.GetClient(),
					Log:         ctrl.Log.WithName("controllers").WithName("Alert"),
					Scheme:      mgr.GetScheme(), // we've added this ourselves
					IsOpenShift: isOpenShift,
				}).SetupWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create controller", "controller", "Alert")
					os.Exit(1)
				}
			case utils.BlackDuckCRDName:
				// setting up Black Duck Controller
				if err = (&controllers.BlackduckReconciler{
					Client:      mgr.GetClient(),
					Log:         ctrl.Log.WithName("controllers").WithName("Black Duck"),
					Scheme:      mgr.GetScheme(), // we've added this ourselves
					IsOpenShift: isOpenShift,
				}).SetupWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create controller", "controller", "Black Duck")
					os.Exit(1)
				}
			case utils.OpsSightCRDName:
				// setting up OpsSight Controller
				if err = (&controllers.OpsSightReconciler{
					Client:      mgr.GetClient(),
					Log:         ctrl.Log.WithName("controllers").WithName("OpsSight"),
					Scheme:      mgr.GetScheme(), // we've added this ourselves
					IsOpenShift: isOpenShift,
				}).SetupWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create controller", "controller", "OpsSight")
					os.Exit(1)
				}
				// setting up OpsSightBlackDuckReconciler Controller
				if err = (&controllers.OpsSightBlackDuckReconciler{
					Client: mgr.GetClient(),
					Log:    ctrl.Log.WithName("controllers").WithName("OpsSightBlackDuck"),
					Scheme: mgr.GetScheme(), // we've added this ourselves
					// TODO: [senthil] add IsOpenShift here?
				}).SetupWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create controller", "controller", "OpsSight")
					os.Exit(1)
				}
			case utils.PolarisCRDName:
				// setting up Polaris Controller
				if err = (&controllers.PolarisReconciler{
					Client:      mgr.GetClient(),
					Log:         ctrl.Log.WithName("controllers").WithName("Polaris"),
					Scheme:      mgr.GetScheme(), // we've added this ourselves
					IsOpenShift: isOpenShift,
				}).SetupWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create controller", "controller", "Polaris")
					os.Exit(1)
				}
				// setting up PolarisDB Controller
				if err = (&controllers.PolarisDBReconciler{
					Client:      mgr.GetClient(),
					Log:         ctrl.Log.WithName("controllers").WithName("PolarisDB"),
					Scheme:      mgr.GetScheme(),
					IsOpenShift: isOpenShift,
				}).SetupWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create controller", "controller", "PolarisDB")
					os.Exit(1)
				}
				if err = (&controllers.AuthServerReconciler{
					Client: mgr.GetClient(),
					Log:    ctrl.Log.WithName("controllers").WithName("AuthServer"),
					Scheme: mgr.GetScheme(),
				}).SetupWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create controller", "controller", "AuthServer")
					os.Exit(1)
				}
			case utils.ReportingCRDName:
				// setting up Reporting Controller
				if err = (&controllers.ReportingReconciler{
					Client:      mgr.GetClient(),
					Log:         ctrl.Log.WithName("controllers").WithName("Reporting"),
					Scheme:      mgr.GetScheme(), // we've added this ourselves
					IsOpenShift: isOpenShift,
				}).SetupWithManager(mgr); err != nil {
					setupLog.Error(err, "unable to create controller", "controller", "Reporting")
					os.Exit(1)
				}
			default:
				fmt.Printf("no controller available for crd %s", crd)
			}
		}
	}
	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
