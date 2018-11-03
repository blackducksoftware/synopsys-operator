package hubutils

import (
	"fmt"

	"github.com/blackducksoftware/perceptor-protoform/pkg/api/hub/v1"
	hubClient "github.com/blackducksoftware/perceptor-protoform/pkg/hub/client/clientset/versioned"
	"github.com/blackducksoftware/perceptor-protoform/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// GetDefaultPasswords returns admin,user,postgres passwords for db maintainance tasks.  Should only be used during
// initialization, or for 'babysitting' ephemeral hub instances (which might have postgres restarts)
// MAKE SURE YOU SEND THE NAMESPACE OF THE SECRET SOURCE (operator), NOT OF THE new hub  THAT YOUR TRYING TO CREATE !
func GetDefaultPasswords(kubeClient *kubernetes.Clientset, nsOfSecretHolder string) (adminPassword string, userPassword string, postgresPassword string, err error) {
	blackduckSecret, err := util.GetSecret(kubeClient, nsOfSecretHolder, "blackduck-secret")
	if err != nil {
		logrus.Infof("warning: You need to first create a 'blackduck-secret' in this namespace with ADMIN_PASSWORD, USER_PASSWORD, POSTGRES_PASSWORD")
		return "", "", "", err
	}
	adminPassword = string(blackduckSecret.Data["ADMIN_PASSWORD"])
	userPassword = string(blackduckSecret.Data["USER_PASSWORD"])
	postgresPassword = string(blackduckSecret.Data["POSTGRES_PASSWORD"])

	// default named return
	return adminPassword, userPassword, postgresPassword, err
}

func updateHubObject(h *hubClient.Clientset, obj *v1.Hub) (*v1.Hub, error) {
	return h.SynopsysV1().Hubs(obj.Name).Update(obj)
}

// UpdateState will be used to update the hub object
func UpdateState(h *hubClient.Clientset, specState string, statusState string, err error, hub *v1.Hub) (*v1.Hub, error) {
	hub.Spec.State = specState
	hub.Status.State = statusState
	if err != nil {
		hub.Status.ErrorMessage = fmt.Sprintf("%+v", err)
	}
	hub, err = updateHubObject(h, hub)
	if err != nil {
		logrus.Errorf("couldn't update the state of hub object: %s", err.Error())
	}
	return hub, err
}
