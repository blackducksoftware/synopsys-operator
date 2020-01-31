package bdba

// BDBA db related constants

// GetBDBADefault returns BDBADB default configuration
func GetBDBADefault() *BDBA {
	return &BDBA{
		Name:                  "bdba",
		Namespace:             "bdba",
		Version:               "20191115",

		// Storage
		PSQLStorageClass: 	   "",
		PSQLSize: 			   "300Gi",
		PSQLExistingClaim: 	   "",
		MinioStorageClass: 	   "",
		MinioSize: 			   "300Gi",
		MinioExistingClaim:    "",
		RabbitMQStorageClass:  "",
		RabbitMQSize: 		   "8Gi",
		RabbitMQExistingClaim: "",

		// Web frontend configuration
		SessionCookieAge:      1209600,
		FrontendReplicas:	   1,
		HideLicenses:          false,
		OfflineMode:		   false,
		AdminEmail:			   "admin@bdba.local",
		ErrorAdminEmail:	   "",
		RootURL:			   "http://bdba.local",

		// SMTP configuration
		EmailEnabled: 		   false,
		EmailSMTPHost:		   "",
		EmailSMTPPort: 		   25,
		EmailSMTPUser: 		   "",
		EmailSMTPPassword: 	   " ",
		EmailFrom: 			   "",
		EmailSecurity: 		   "none",
		EmailVerify: 		   false,

		// LDAP
		LDAPEnabled: 		  false,
		LDAPServerURI: 		  "",
		LDAPUserDNTemplate:   "",
		LDAPBindAsAuthenticating: true,
		LDAPBindDN: 		  "",
		LDAPBindPassword: 	  " ",
		LDAPStartTLS: 		  false,
		LDAPVerify: 		  false,
		LDAPRootCASecret: 	  "",
		LDAPRootCAFile: 	  "",
		LDAPRequireGroup: 	  "",
		LDAPUserSearch: 	  "",
		LDAPUserSearchScope:  "",
		LDAPGroupSearch: 	  "",
		LDAPGroupSearchScope: "",
		LDAPNestedSearch: 	  false,

		// Licensing
		LicensingUsername:     "",
		LicensingPassword:     "",
		LicensingUpstream:	   "https://protecode-sc.com",

		// Logging
		FrontendLogging: 	   true,
		WorkerLogging: 		   true,

		// Worker scaling
		WorkerReplicas:        1,
		WorkerConcurrency:	   1,

		// Networking and security
		RootCASecret: 		   "",
		HTTPProxy: 			   "",
		HTTPNoProxy: 		   "",

		// Ingress
		IngressEnabled:		   true,
		IngressHost:		   "bdba.local",
		IngressTLSEnabled:	   false,
		IngressTLSSecretName:  "bdba-ingress-tls",

		// RabbitMQ
		BrokerURL:             "amqp://bdba:%s@bdba-rabbitmq",
		RabbitMQK8SDomain:	   "protecode-sc-cluster.local",

		// External PG
		PGHost:				   "",
		PGPort:				   "",
		PGUser:				   "",
		PGDataBase:			   "",
	}
}
