package bdba

// BDBA db related constants

const (
	// SOMETHING is something like the size of a database
	SOMETHING = "10Gi"
)

// GetBDBADefault returns BDBADB default configuration
func GetBDBADefault() *BDBA {
	return &BDBA{
		Name:                  "namedefault",
		Namespace:             "namespacedefault",
		Version:               "20191115",
		Hostname:              "hostnamedefault",
		IngressHost:           "ingressHost.default",
		MinioAccessKey:        "QUJDREVGRzEyMzQ1Njc=",
		MinioSecretKey:        "QUJDREVGRzEyMzQ1Njc=",
		WorkerReplicas:        1234,
		AdminEmail:            "adminEmaildefault",
		BrokerURL:             "QUJDREVGRzEyMzQ1Njc=",
		PGPPassword:           "QUJDREVGRzEyMzQ1Njc=",
		RabbitMQULimitNoFiles: "rabbitMQULimitNoFilesdefault",
		HideLicenses:          "hideLicensesdefault",
		LicensingPassword:     "QUJDREVGRzEyMzQ1Njc=",
		LicensingUsername:     "QUJDREVGRzEyMzQ1Njc=",
		InsecureCookies:       "insecureCookiesdefault",
		SessionCookieAge:      "sessionCookieAgedefault",
		URL:                   "urldefault",
		Actual:                "actualdefault",
		Expected:              "expecteddefault",
		StartFlag:             "startFlagdefault",
		Result:                "resultdefault",
	}
}
