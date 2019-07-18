package v1

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	utils2 "github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/components/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type BdRSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func (b BdRSecret) GetSecrets() []*components.Secret {
	var secrets []*components.Secret

	if len(b.blackduck.Spec.AuthCustomCA) > 0 {
		cert, err := stringToCertificate(b.blackduck.Spec.AuthCustomCA)
		if err != nil {
			logrus.Warnf("The Auth Custom CA provided is invalid")
		} else {
			logrus.Debugf("Adding The Auth Custom CA with SN: %x", cert.SerialNumber)
			authCustomCASecret := components.NewSecret(horizonapi.SecretConfig{Namespace: b.blackduck.Spec.Namespace, Name: util.GetResourceName(b.blackduck.Name, util.BlackDuckName, "auth-custom-ca"), Type: horizonapi.SecretTypeOpaque})
			authCustomCASecret.AddData(map[string][]byte{"AUTH_CUSTOM_CA": []byte(b.blackduck.Spec.AuthCustomCA)})
			authCustomCASecret.AddLabels(utils2.GetVersionLabel("secret", b.blackduck.Name, b.blackduck.Spec.Version))
			secrets = append(secrets, authCustomCASecret)
		}
	}
	return secrets
}

func stringToCertificate(certificate string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certificate))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}

	return cert, nil
}
func NewBdRSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.SecretInterface {
	return &BdRSecret{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func init() {
	store.Register(types.SecretAuthCertificateV1, NewBdRSecret)
}
