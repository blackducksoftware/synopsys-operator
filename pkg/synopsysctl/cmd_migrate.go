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

package synopsysctl

import (
	"fmt"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// migrateCmd migrates a resource before upgrading Synopsys Operator
var migrateCmd = &cobra.Command{
	Use:           "migrate",
	Example:       "synopsysctl migrate <from>\nsynopsysctl migrate 2019.4.2",
	Short:         "Migrate a Synopsys resource before upgrading the operator",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		// Check number of arguments
		if len(args) != 1 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(namespace) > 0 {
			return migrate(namespace)
		}

		// get operator namespace
		isClusterScoped := util.GetClusterScope(apiExtensionClient)
		if isClusterScoped {
			namespaces, err := util.GetOperatorNamespace(kubeClient, metav1.NamespaceAll)
			if err != nil {
				return err
			}

			if len(namespaces) > 1 {
				return fmt.Errorf("more than 1 Synopsys Operator found in your cluster. please pass the namespace of the Synopsys Operator that you want to migrate")
			}
			return migrate(namespaces[0])
		}

		log.Errorf("namespace of the Synopsys Operator need to be provided")
		return nil
	},
}

func migrate(namespace string) error {
	err := migrateCRD(namespace)
	if err != nil {
		return err
	}
	err = migrateCR(namespace)
	if err != nil {
		return err
	}
	return nil
}

// migrateCRD adds the labels to the custom resource definitions for the existing operator
func migrateCRD(namespace string) error {
	crdNames := []string{util.AlertCRDName, util.BlackDuckCRDName, util.OpsSightCRDName}
	for _, crdName := range crdNames {
		crd, err := util.GetCustomResourceDefinition(apiExtensionClient, crdName)
		if err != nil {
			return fmt.Errorf("error getting %s custom resource defintion due to %+v", crdName, err)
		}

		// if crd labels doesn't contain app, then updates
		if _, ok := crd.Labels[fmt.Sprintf("synopsys.com/operator.%s", namespace)]; !ok {
			crd.Labels = util.InitLabels(crd.Labels)
			crd.Labels["app"] = "synopsys-operator"
			crd.Labels["component"] = "operator"
			crd.Labels[fmt.Sprintf("synopsys.com/operator.%s", namespace)] = namespace
			_, err = util.UpdateCustomResourceDefinition(apiExtensionClient, crd)
			if err != nil {
				return fmt.Errorf("unable to update %s custom resource defintion due to %+v", crdName, err)
			}
		}
		log.Infof("successfully migrated '%s' custom resource definition", crd.GetName())
	}
	return nil
}

// migrateCR add the labels to the existing custom resource instances
func migrateCR(namespace string) error {
	crdNames := []string{util.AlertCRDName, util.BlackDuckCRDName, util.OpsSightCRDName}
	for _, crdName := range crdNames {
		switch crdName {
		case util.AlertCRDName:
			err := migrateAlert(namespace)
			if err != nil {
				return err
			}
		case util.BlackDuckCRDName:
			err := migrateBlackDuck(namespace)
			if err != nil {
				return err
			}
		case util.OpsSightCRDName:
			err := migrateOpsSight(namespace)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// migrateAlert migrates the existing Alert instances
func migrateAlert(namespace string) error {
	alerts, err := util.ListAlerts(alertClient, namespace)
	if err != nil {
		return fmt.Errorf("failed to list Alert instances in namespace '%s' due to %+v", namespace, err)
	}
	for _, alert := range alerts.Items {
		alertName := alert.Name
		alertNamespace := alert.Spec.Namespace
		log.Infof("migrating Alert '%s' in namespace '%s'...", alertName, alertNamespace)

		// update annotations
		if _, ok := alert.Annotations["synopsys.com/created.by"]; !ok {
			alert.Annotations = util.InitAnnotations(alert.Annotations)
			alert.Annotations["synopsys.com/created.by"] = "pre-2019.6.0"
			_, err := alertClient.SynopsysV1().Alerts(alertNamespace).Update(&alert)
			if err != nil {
				return fmt.Errorf("error migrating Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
			}
		}

		// add synopsys labels to namespace
		err = addNamespaceLabels(alertNamespace, alertName, util.AlertName, alert.Spec.Version)
		if err != nil {
			return err
		}

		// include name in all resources
		err = addNameLabels(alertNamespace, alertName, util.AlertName)
		if err != nil {
			return err
		}
		log.Infof("successfully migrated Alert '%s' in namespace '%s'", alertName, alertNamespace)
	}
	return nil
}

// migrateBlackDuck migrates the existing Black Duck instances
func migrateBlackDuck(namespace string) error {
	blackDucks, err := util.ListHubs(blackDuckClient, namespace)
	if err != nil {
		return fmt.Errorf("failed to list Black Duck instances in namespace '%s' due to %+v", namespace, err)
	}
	for _, blackDuck := range blackDucks.Items {
		blackDuckName := blackDuck.Name
		blackDuckNamespace := blackDuck.Spec.Namespace
		log.Infof("migrating Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)

		// ASSUMING ALL PASSWORDS HAVE REMAINED THE SAME, no need to pull from secret
		defaultPassword := util.Base64Encode([]byte("blackduck"))

		blackDuck.Spec.AdminPassword = defaultPassword
		blackDuck.Spec.UserPassword = defaultPassword
		blackDuck.Spec.PostgresPassword = defaultPassword

		// update annotations
		if _, ok := blackDuck.Annotations["synopsys.com/created.by"]; !ok {
			blackDuck.Annotations = util.InitAnnotations(blackDuck.Annotations)
			blackDuck.Annotations["synopsys.com/created.by"] = "pre-2019.6.0"
			_, err := blackDuckClient.SynopsysV1().Blackducks(blackDuckNamespace).Update(&blackDuck)
			if err != nil {
				return fmt.Errorf("error migrating Black Duck '%s' in namespace '%s' due to %+v", blackDuckName, blackDuckNamespace, err)
			}
		}

		// add synopsys labels to namespace
		err = addNamespaceLabels(blackDuckNamespace, blackDuckName, util.BlackDuckName, blackDuck.Spec.Version)
		if err != nil {
			return err
		}

		// include name in all resources
		err = addNameLabels(blackDuckNamespace, blackDuckName, util.BlackDuckName)
		if err != nil {
			return err
		}
		log.Infof("successfully migrated Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
	}
	return nil
}

// migrateOpsSight migrates the existing OpsSight instances
func migrateOpsSight(namespace string) error {
	opsSights, err := util.ListOpsSights(opsSightClient, namespace)
	if err != nil {
		return fmt.Errorf("failed to list OpsSight instances in namespace '%s' due to %+v", namespace, err)
	}
	for _, opsSight := range opsSights.Items {
		opsSightName := opsSight.Name
		opsSightNamespace := opsSight.Spec.Namespace
		log.Infof("migrating OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)

		// set the desired state to empty which is equivalent to START
		opsSight.Spec.DesiredState = ""

		// ASSUMING ALL PASSWORDS HAVE REMAINED THE SAME, no need to pull from secret
		defaultPassword := util.Base64Encode([]byte("blackduck"))
		opsSight.Spec.Blackduck.BlackduckPassword = defaultPassword
		opsSight.Spec.Blackduck.BlackduckSpec.AdminPassword = defaultPassword
		opsSight.Spec.Blackduck.BlackduckSpec.UserPassword = defaultPassword
		opsSight.Spec.Blackduck.BlackduckSpec.PostgresPassword = defaultPassword

		// update annotations
		if _, ok := opsSight.Annotations["synopsys.com/created.by"]; !ok {
			opsSight.Annotations = util.InitAnnotations(opsSight.Annotations)
			opsSight.Annotations["synopsys.com/created.by"] = "pre-2019.6.0"
			_, err := opsSightClient.SynopsysV1().OpsSights(opsSightNamespace).Update(&opsSight)
			if err != nil {
				return fmt.Errorf("error migrating OpsSight '%s' in namespace '%s' due to %+v", opsSightName, opsSightNamespace, err)
			}
		}

		// add synopsys labels to namespace
		err = addNamespaceLabels(opsSightNamespace, opsSightName, util.OpsSightName, "2.2.3")
		if err != nil {
			return err
		}

		// include name in all resources
		err = addNameLabels(opsSightNamespace, opsSightName, util.BlackDuckName)
		if err != nil {
			return err
		}
		log.Infof("successfully migrated OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
	}
	return nil
}

// addNamespaceLabels adds a synopsys label to the namespace label
func addNamespaceLabels(namespace string, name string, appName string, version string) error {
	ns, err := util.GetNamespace(kubeClient, namespace)
	if err != nil {
		return fmt.Errorf("error getting %s namespace due to %+v", namespace, err)
	}

	// update labels in namespace
	if _, ok := ns.Labels[fmt.Sprintf("synopsys.com/%s.%s", appName, name)]; !ok {
		ns.Labels = util.InitLabels(ns.Labels)
		ns.Labels["owner"] = util.OperatorName
		ns.Labels[fmt.Sprintf("synopsys.com/%s.%s", appName, name)] = version
		_, err = util.UpdateNamespace(kubeClient, ns)
		if err != nil {
			return fmt.Errorf("error updating %s namespace due to %+v", namespace, err)
		}
	}
	return nil
}

// addNameLabels adds a name label to all its resources
func addNameLabels(namespace string, name string, appName string) error {
	deployments, err := util.ListDeployments(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
	if err != nil {
		return fmt.Errorf("unable to list deployments for %s %s in namespace %s due to %+v", appName, name, namespace, err)
	}

	for _, deployment := range deployments.Items {
		if _, ok := deployment.Labels["name"]; !ok {
			deployment.Labels = util.InitLabels(deployment.Labels)
			deployment.Labels["name"] = name
			deployment.Spec.Template.Labels = util.InitLabels(deployment.Spec.Template.Labels)
			deployment.Spec.Template.Labels["name"] = name
			_, err = util.UpdateDeployment(kubeClient, namespace, &deployment)
			if err != nil {
				return fmt.Errorf("unable to update %s deployment in namespace %s due to %+v", deployment.GetName(), namespace, err)
			}
		}
	}

	rcs, err := util.ListReplicationControllers(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
	if err != nil {
		return fmt.Errorf("unable to list replication controllers for %s %s in namespace %s due to %+v", appName, name, namespace, err)
	}

	for _, rc := range rcs.Items {
		if _, ok := rc.Labels["name"]; !ok {
			rc.Labels = util.InitLabels(rc.Labels)
			rc.Labels["name"] = name
			rc.Spec.Template.Labels = util.InitLabels(rc.Spec.Template.Labels)
			rc.Spec.Template.Labels["name"] = name
			_, err = util.UpdateReplicationController(kubeClient, namespace, &rc)
			if err != nil {
				return fmt.Errorf("unable to update %s replication controller in namespace %s due to %+v", rc.GetName(), namespace, err)
			}
		}
	}

	services, err := util.ListServices(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
	if err != nil {
		return fmt.Errorf("unable to list services for %s %s in namespace %s due to %+v", appName, name, namespace, err)
	}

	for _, svc := range services.Items {
		if _, ok := svc.Labels["name"]; !ok {
			svc.Labels = util.InitLabels(svc.Labels)
			svc.Labels["name"] = name
			_, err = util.UpdateService(kubeClient, namespace, &svc)
			if err != nil {
				return fmt.Errorf("unable to update %s service in namespace %s due to %+v", svc.GetName(), namespace, err)
			}
		}
	}

	configmaps, err := util.ListConfigMaps(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
	if err != nil {
		return fmt.Errorf("unable to list config maps for %s %s in namespace %s due to %+v", appName, name, namespace, err)
	}

	for _, cm := range configmaps.Items {
		if _, ok := cm.Labels["name"]; !ok {
			cm.Labels = util.InitLabels(cm.Labels)
			cm.Labels["name"] = name
			_, err = util.UpdateConfigMap(kubeClient, namespace, &cm)
			if err != nil {
				return fmt.Errorf("unable to update %s config map in namespace %s due to %+v", cm.GetName(), namespace, err)
			}
		}
	}

	secrets, err := util.ListSecrets(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
	if err != nil {
		return fmt.Errorf("unable to list secrets for %s %s in namespace %s due to %+v", appName, name, namespace, err)
	}

	for _, secret := range secrets.Items {
		if _, ok := secret.Labels["name"]; !ok {
			secret.Labels = util.InitLabels(secret.Labels)
			secret.Labels["name"] = name
			_, err = util.UpdateSecret(kubeClient, namespace, &secret)
			if err != nil {
				return fmt.Errorf("unable to update %s secret in namespace %s due to %+v", secret.GetName(), namespace, err)
			}
		}
	}

	pvcs, err := util.ListPVCs(kubeClient, namespace, fmt.Sprintf("app=%s", appName))
	if err != nil {
		return fmt.Errorf("unable to list persistent volume claims for %s %s in namespace %s due to %+v", appName, name, namespace, err)
	}

	for _, pvc := range pvcs.Items {
		if _, ok := pvc.Labels["name"]; !ok {
			pvc.Labels = util.InitLabels(pvc.Labels)
			pvc.Labels["name"] = name
			_, err = util.UpdatePVC(kubeClient, namespace, &pvc)
			if err != nil {
				return fmt.Errorf("unable to update %s persistent volume claim in namespace %s due to %+v", pvc.GetName(), namespace, err)
			}
		}
	}

	routeClient := util.GetRouteClient(restconfig, kubeClient, namespace)
	if routeClient != nil {
		routes, err := util.ListRoutes(routeClient, namespace, fmt.Sprintf("app=%s", appName))
		if err != nil {
			return fmt.Errorf("unable to list routes for %s %s in namespace %s due to %+v", appName, name, namespace, err)
		}

		for _, route := range routes.Items {
			if _, ok := route.Labels["name"]; !ok {
				route.Labels = util.InitLabels(route.Labels)
				route.Labels["name"] = name
				_, err = util.UpdateRoute(routeClient, namespace, &route)
				if err != nil {
					return fmt.Errorf("unable to update %s route in namespace %s due to %+v", route.GetName(), namespace, err)
				}
			}
		}
	}

	return nil
}

func init() {
	// Add Migrate Commands
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
}
