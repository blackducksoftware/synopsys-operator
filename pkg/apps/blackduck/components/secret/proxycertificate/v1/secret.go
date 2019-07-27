package v1

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/types"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// BdRSecret holds the Black Duck secret configuration
type BdRSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackDuck  *blackduckapi.Blackduck
}

// GetSecrets returns the secret
func (b BdRSecret) GetSecrets() []*components.Secret {
	var secrets []*components.Secret
	if len(b.blackDuck.Spec.ProxyCertificate) > 0 {
		cert, err := stringToCertificate(b.blackDuck.Spec.ProxyCertificate)
		if err != nil {
			logrus.Warnf("The proxy certificate provided is invalid")
		} else {
			logrus.Debugf("Adding Proxy certificate with SN: %x", cert.SerialNumber)
			proxyCertificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: b.blackDuck.Spec.Namespace, Name: apputils.GetResourceName(b.blackDuck.Name, util.BlackDuckName, "proxy-certificate"), Type: horizonapi.SecretTypeOpaque})
			proxyCertificateSecret.AddData(map[string][]byte{"HUB_PROXY_CERT_FILE": []byte(b.blackDuck.Spec.ProxyCertificate)})
			proxyCertificateSecret.AddLabels(apputils.GetVersionLabel("secret", b.blackDuck.Name, b.blackDuck.Spec.Version))
			secrets = append(secrets, proxyCertificateSecret)
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

// NewBdRSecret returns the Black Duck secret configuration
func NewBdRSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, cr interface{}) (types.SecretInterface, error) {
	blackDuck, ok := cr.(*blackduckapi.Blackduck)
	if !ok {
		return nil, fmt.Errorf("unable to cast the interface to Black Duck object")
	}
	return &BdRSecret{config: config, kubeClient: kubeClient, blackDuck: blackDuck}, nil
}

func init() {
	store.Register(blackduck.BlackDuckProxyCertificateSecretV1, NewBdRSecret)
}
