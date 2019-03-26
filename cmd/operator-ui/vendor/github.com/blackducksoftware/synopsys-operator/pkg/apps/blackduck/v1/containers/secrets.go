package containers

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	"github.com/sirupsen/logrus"
	"strings"
)

func (hc *Creater) GetSecrets( adminPassword string, userPassword string) []*components.Secret {
	var secrets []*components.Secret
	hubSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: hc.hubSpec.Namespace, Name: "db-creds", Type: horizonapi.SecretTypeOpaque})

	if hc.hubSpec.ExternalPostgres != nil {
		hubSecret.AddStringData(map[string]string{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": hc.hubSpec.ExternalPostgres.PostgresAdminPassword, "HUB_POSTGRES_USER_PASSWORD_FILE": hc.hubSpec.ExternalPostgres.PostgresUserPassword})
	} else {
		hubSecret.AddStringData(map[string]string{"HUB_POSTGRES_ADMIN_PASSWORD_FILE": adminPassword, "HUB_POSTGRES_USER_PASSWORD_FILE": userPassword})
	}
	secrets = append(secrets, hubSecret)

	certificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: hc.hubSpec.Namespace, Name: "blackduck-certificate", Type: horizonapi.SecretTypeOpaque})
	if strings.EqualFold(hc.hubSpec.CertificateName, "manual") {
		certificateSecret.AddData(map[string][]byte{"WEBSERVER_CUSTOM_CERT_FILE": []byte(hc.hubSpec.Certificate), "WEBSERVER_CUSTOM_KEY_FILE": []byte(hc.hubSpec.CertificateKey)})
	}
	secrets = append(secrets, certificateSecret)

	if len(hc.hubSpec.ProxyCertificate) > 0 {
		cert, err := hc.stringToCertificate(hc.hubSpec.ProxyCertificate)
		if err != nil {
			logrus.Warnf("The proxy certificate provided is invalid")
		} else {
			logrus.Debugf("Adding Proxy certificate with SN: %x", cert.SerialNumber)
			proxyCertificateSecret := components.NewSecret(horizonapi.SecretConfig{Namespace: hc.hubSpec.Namespace, Name: "blackduck-proxy-certificate", Type: horizonapi.SecretTypeOpaque})
			proxyCertificateSecret.AddData(map[string][]byte{"HUB_PROXY_CERT_FILE": []byte(hc.hubSpec.ProxyCertificate)})
			secrets = append(secrets, proxyCertificateSecret)
		}
	}

	if len(hc.hubSpec.AuthCustomCA) > 0 {
		cert, err := hc.stringToCertificate(hc.hubSpec.AuthCustomCA)
		if err != nil {
			logrus.Warnf("The Auth Custom CA provided is invalid")
		} else {
			logrus.Debugf("Adding The Auth Custom CA with SN: %x", cert.SerialNumber)
			authCustomCASecret := components.NewSecret(horizonapi.SecretConfig{Namespace: hc.hubSpec.Namespace, Name: "auth-custom-ca", Type: horizonapi.SecretTypeOpaque})
			authCustomCASecret.AddData(map[string][]byte{"AUTH_CUSTOM_CA": []byte(hc.hubSpec.AuthCustomCA)})
			secrets = append(secrets, authCustomCASecret)
		}
	}

	return secrets
}

func (hc *Creater) stringToCertificate(certificate string) (*x509.Certificate, error) {
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
