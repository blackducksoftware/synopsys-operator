package v1

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	blackduckapi "github.com/blackducksoftware/synopsys-operator/pkg/api/blackduck/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/blackduck/types"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps/store"
	apputils "github.com/blackducksoftware/synopsys-operator/pkg/apps/utils"
	"github.com/blackducksoftware/synopsys-operator/pkg/protoform"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	"math/big"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
)

type BdRSecret struct {
	config     *protoform.Config
	kubeClient *kubernetes.Clientset
	blackduck  *blackduckapi.Blackduck
}

func (b BdRSecret) GetSecrets() []*components.Secret {
	var secrets []*components.Secret
	certificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: b.blackduck.Spec.Namespace, Name: apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "webserver-certificate"), Type: horizonapi.SecretTypeOpaque})

	cert, key, _ := b.getTLSCertKeyOrCreate()
	certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(cert), "WEBSERVER_CUSTOM_KEY_FILE": []byte(key)})
	certificateSecret.AddLabels(apputils.GetVersionLabel("secret", b.blackduck.Name, b.blackduck.Spec.Version))

	secrets = append(secrets, certificateSecret)
	return secrets
}

func (b BdRSecret) getTLSCertKeyOrCreate() (string, string, error) {
	if len(b.blackduck.Spec.Certificate) > 0 && len(b.blackduck.Spec.CertificateKey) > 0 {
		return b.blackduck.Spec.Certificate, b.blackduck.Spec.CertificateKey, nil
	}

	// Cert copy
	if len(b.blackduck.Spec.CertificateName) > 0 && !strings.EqualFold(b.blackduck.Spec.CertificateName, "default") {
		secret, err := util.GetSecret(b.kubeClient, b.blackduck.Spec.CertificateName, apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "webserver-certificate"))
		if err == nil {
			cert, certok := secret.Data["WEBSERVER_CUSTOM_CERT_FILE"]
			key, keyok := secret.Data["WEBSERVER_CUSTOM_KEY_FILE"]
			if certok && keyok {
				return string(cert), string(key), nil
			}
		}
	}

	// default cert
	secret, err := util.GetSecret(b.kubeClient, b.config.Namespace, "blackduck-certificate")
	if err == nil {
		data := secret.Data
		if len(data) >= 2 {
			cert, certok := secret.Data["WEBSERVER_CUSTOM_CERT_FILE"]
			key, keyok := secret.Data["WEBSERVER_CUSTOM_KEY_FILE"]
			if !certok || !keyok {
				util.DeleteSecret(b.kubeClient, b.blackduck.Spec.Namespace, apputils.GetResourceName(b.blackduck.Name, util.BlackDuckName, "webserver-certificate"))
			} else {
				return string(cert), string(key), nil
			}
		}
	}

	// Default
	return CreateSelfSignedCert()
}

// CreateSelfSignedCert will create a random self signed certificate
func CreateSelfSignedCert() (string, string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", err
	}
	//Max random value, a 130-bits integer, i.e 2^130 - 1
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(130), nil).Sub(max, big.NewInt(1))
	template := x509.Certificate{
		SerialNumber: max,
		Subject: pkix.Name{
			CommonName:         "Black Duck",
			OrganizationalUnit: []string{"Cloud Native"},
			Organization:       []string{"Black Duck By Synopsys"},
			Locality:           []string{"Burlington"},
			StreetAddress:      []string{"800 District Avenue"},
			Province:           []string{"Massachusetts"},
			Country:            []string{"US"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour * 24 * 365 * 3),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		return "", "", err
	}
	certificate := &bytes.Buffer{}
	key := &bytes.Buffer{}
	pem.Encode(certificate, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	pemBlock, err := pemBlockForKey(priv)
	if err != nil {
		return "", "", err
	}

	pem.Encode(key, pemBlock)
	return certificate.String(), key.String(), nil
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func pemBlockForKey(priv interface{}) (*pem.Block, error) {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}, nil
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal ECDSA private key: %v", err)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}, nil
	default:
		return nil, nil
	}
}

func NewBdRSecret(config *protoform.Config, kubeClient *kubernetes.Clientset, blackduck *blackduckapi.Blackduck) types.SecretInterface {
	return &BdRSecret{config: config, kubeClient: kubeClient, blackduck: blackduck}
}

func init() {
	store.Register(types.SecretWebCertificateV1, NewBdRSecret)
}
