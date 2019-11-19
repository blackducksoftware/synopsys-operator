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
	"encoding/json"
	"fmt"
	"io"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"

	"github.com/blackducksoftware/synopsys-operator/pkg/polaris"

	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var promptAnswerYes bool

// deleteCmd deletes a resource from the cluster
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove Synopsys resources from your cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("must specify a sub-command")
	},
}

// deleteAlertCmd deletes Alert instances from the cluster
var deleteAlertCmd = &cobra.Command{
	Use:           "alert NAME...",
	Example:       "synopsysctl delete alert <name>\nsynopsysctl delete alert <name1> <name2> <name3>\nsynopsysctl delete alert <name> -n <namespace>\nsynopsysctl delete alert <name1> <name2> <name3> -n <namespace>",
	Short:         "Delete one or many Alert instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, alertName := range args {
			alertNamespace, crdNamespace, _, err := getInstanceInfo(false, util.AlertCRDName, util.AlertName, namespace, alertName)
			if err != nil {
				return err
			}
			log.Infof("deleting Alert '%s' in namespace '%s'...", alertName, alertNamespace)
			err = util.DeleteAlert(alertClient, alertName, crdNamespace, &metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting Alert '%s' in namespace '%s' due to %+v", alertName, alertNamespace, err)
			}
			log.Infof("successfully submitted delete Alert '%s' in namespace '%s'", alertName, alertNamespace)
		}
		return nil
	},
}

// deleteBlackDuckCmd deletes Black Duck instances from the cluster
var deleteBlackDuckCmd = &cobra.Command{
	Use:           "blackduck NAME...",
	Example:       "synopsysctl delete blackduck <name>\nsynopsysctl delete blackduck <name1> <name2> <name3>\nsynopsysctl delete blackduck <name> -n <namespace>\nsynopsysctl delete blackduck <name1> <name2> <name3> -n <namespace>",
	Short:         "Delete one or many Black Duck instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, blackDuckName := range args {
			blackDuckNamespace, crdNamespace, _, err := getInstanceInfo(false, util.BlackDuckCRDName, util.BlackDuckName, namespace, blackDuckName)
			if err != nil {
				return err
			}
			log.Infof("deleting Black Duck '%s' in namespace '%s'...", blackDuckName, blackDuckNamespace)
			err = util.DeleteBlackduck(blackDuckClient, blackDuckName, crdNamespace, &metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting Black Duck '%s' in namespace '%s' due to '%s'", blackDuckName, blackDuckNamespace, err)
			}
			log.Infof("successfully submitted delete Black Duck '%s' in namespace '%s'", blackDuckName, blackDuckNamespace)
		}
		return nil
	},
}

// deleteOpsSightCmd deletes OpsSight instances from the cluster
var deleteOpsSightCmd = &cobra.Command{
	Use:           "opssight NAME...",
	Example:       "synopsysctl delete opssight <name>\nsynopsysctl delete opssight <name1> <name2> <name3>\nsynopsysctl delete opssight <name> -n <namespace>\nsynopsysctl delete opssight <name1> <name2> <name3> -n <namespace>",
	Short:         "Delete one or many OpsSight instances",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			cmd.Help()
			return fmt.Errorf("this command takes 1 or more arguments")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, opsSightName := range args {
			opsSightNamespace, crdNamespace, _, err := getInstanceInfo(false, util.OpsSightCRDName, util.OpsSightName, namespace, opsSightName)
			if err != nil {
				return err
			}
			log.Infof("deleting OpsSight '%s' in namespace '%s'...", opsSightName, opsSightNamespace)
			err = util.DeleteOpsSight(opsSightClient, opsSightName, crdNamespace, &metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("error deleting OpsSight '%s' in namespace '%s' due to '%s'", opsSightName, opsSightNamespace, err)
			}
			log.Infof("successfully submitted delete OpsSight '%s' in namespace '%s'", opsSightName, opsSightNamespace)
		}
		return nil
	},
}

// deletePolarisCmd deletes Polaris instances from the cluster
var deletePolarisCmd = &cobra.Command{
	Use:           "polaris",
	Example:       "synopsysctl delete polaris -n <namespace>",
	Short:         "Delete a polaris instance",
	SilenceUsage:  true,
	SilenceErrors: true,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			cmd.Help()
			return fmt.Errorf("this command takes no argument")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(namespace) == 0 {
			return fmt.Errorf("a namespace must be provided using -n")
		}

		p, err := getPolarisFromSecret()
		if err != nil {
			return err
		}

		if p == nil {
			return fmt.Errorf("either namespace does not exist or secret does not exist because this instance of polaris was not created via synopsysctl")
		}

		components, err := polaris.GetComponents(baseURL, *p)
		if err != nil {
			return err
		}

		var content []byte
		for _, v := range components {
			// We skip it if it's a PVC
			if _, ok := v.(*v1.PersistentVolumeClaim); ok {
				continue
			}

			polarisComponentsByte, err := json.Marshal(v)
			if err != nil {
				return err
			}
			content = append(content, polarisComponentsByte...)
		}

		log.Info("Deleting Polaris")

		// Delete components from the yaml file
		out, err := RunKubeCmdWithStdin(restconfig, kubeClient, string(content), "delete", "-f", "-")
		if err != nil {
			continueAnswer, err := AskYesNoWithDefault(func() error {
				log.Warn("Some errors occurred during the deletion of Polaris component. Do you want to continue? (Yes/No)")
				return nil
			}, promptAnswerYes, os.Stdin)

			if err != nil {
				return err
			}

			if !continueAnswer {
				return fmt.Errorf("couldn't delete polaris |  %+v - %s", out, err)
			}
		}

		// Delete secret generated by init jobs
		if list, err := kubeClient.CoreV1().Secrets(namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("environment=%s", p.Namespace)}); err == nil {
			for _, v := range list.Items {
				if err := kubeClient.CoreV1().Secrets(namespace).Delete(v.Name, &metav1.DeleteOptions{}); err != nil {
					log.Warnf("Couldn't delete secret %s in namespace %s", v.Name, namespace)
				}
			}
		}

		// Delete StatefulSet PVCs
		if list, err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{LabelSelector: fmt.Sprintf("environment=%s", p.Namespace)}); err == nil {
			if len(list.Items) > 0 {

				deletePvc, err := AskYesNoWithDefault(func() error {
					log.Warn("Do you want to delete the following PVCs? (Yes/No)")
					for _, v := range list.Items {
						log.Warn(v.Name)
					}
					return nil
				}, promptAnswerYes, os.Stdin)

				if err != nil {
					return err
				}

				if deletePvc {
					for _, v := range list.Items {
						log.Warnf("deleting %s", v.Name)
						if err := kubeClient.CoreV1().PersistentVolumeClaims(namespace).Delete(v.Name, &metav1.DeleteOptions{}); err != nil {
							log.Warnf("Couldn't delete PVC %s in namespace %s", v.Name, namespace)
						}
					}
				}
			}
		}

		// Delete the polaris secret that contains the configuration
		if err := kubeClient.CoreV1().Secrets(namespace).Delete("polaris", &metav1.DeleteOptions{}); err != nil {
			return err
		}

		log.Info("Polaris has been successfully Deleted!")
		return nil
	},
}

func AskYesNoWithDefault(question func() error, alwaysYes bool, reader io.Reader) (bool, error) {
	if alwaysYes {
		return alwaysYes, nil
	}
	if err := question(); err != nil {
		return false, err
	}
	return askYesNo(reader)
}

func askYesNo(reader io.Reader) (bool, error) {
	var resp string
	fmt.Fscanln(reader, &resp)

	if strings.EqualFold("yes", strings.TrimSpace(resp)) {
		return true, nil
	} else if strings.EqualFold("no", strings.TrimSpace(resp)) {
		return false, nil
	} else {
		fmt.Println("Invalid response. Please enter either yes or no: ")
		return askYesNo(reader)
	}
}

func init() {
	//(PassCmd) deleteCmd.DisableFlagParsing = true // lets deleteCmd pass flags to kube/oc
	rootCmd.AddCommand(deleteCmd)

	// Add Delete Alert Command
	deleteAlertCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	deleteCmd.AddCommand(deleteAlertCmd)

	// Add Delete Black Duck Command
	deleteBlackDuckCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	deleteCmd.AddCommand(deleteBlackDuckCmd)

	// Add Delete OpsSight Command
	deleteOpsSightCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	deleteCmd.AddCommand(deleteOpsSightCmd)

	// Add Delete Polaris Command
	deletePolarisCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the instance(s)")
	deletePolarisCmd.Flags().BoolVarP(&promptAnswerYes, "yes", "y", promptAnswerYes, "Automatic yes to prompts")
	addbaseURLFlag(deletePolarisCmd)
	deleteCmd.AddCommand(deletePolarisCmd)
}
