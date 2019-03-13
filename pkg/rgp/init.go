package rgp

import (
	"database/sql"
	"fmt"
	"time"

	horizonapi "github.com/blackducksoftware/horizon/pkg/api"
	"github.com/blackducksoftware/horizon/pkg/components"
	deployer2 "github.com/blackducksoftware/horizon/pkg/deployer"
	"github.com/blackducksoftware/synopsys-operator/pkg/api/rgp/v1"
	"github.com/blackducksoftware/synopsys-operator/pkg/apps"
	"github.com/blackducksoftware/synopsys-operator/pkg/util"
	log "github.com/sirupsen/logrus"
	v1_batch "k8s.io/api/batch/v1"
	v14 "k8s.io/api/core/v1"
	v12 "k8s.io/api/rbac/v1"
	v13 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// init deploys  minio, vault and consul
func (c *Creater) init(spec *v1.RgpSpec) error {
	const vaultConfig = `{"listener":{"tcp":{"address":"[::]:8200","cluster_address":"[::]:8201","tls_cert_file":"/vault/tls/tls.crt","tls_cipher_suites":"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,TLS_RSA_WITH_AES_128_GCM_SHA256,TLS_RSA_WITH_AES_256_GCM_SHA384,TLS_RSA_WITH_AES_128_CBC_SHA,TLS_RSA_WITH_AES_256_CBC_SHA","tls_disable":false,"tls_key_file":"/vault/tls/tls.key","tls_prefer_server_cipher_suites":true}},"storage":{"consul":{"address":"consul:8500","path":"vault"}}}`

	err := c.eventStoreInit(spec)

	// Minio
	minioclaim, _ := util.CreatePersistentVolumeClaim("minio", spec.Namespace, "1Gi", spec.StorageClass, horizonapi.ReadWriteOnce)
	minioDeployer, _ := deployer2.NewDeployer(c.kubeConfig)

	// TODO generate random password
	minioCreater := apps.NewMinio(spec.Namespace, "minio", "aaaa2wdadwdawdawd", "b2112r43rfefefbbb")
	minioDeployer.AddSecret(minioCreater.GetSecret())
	minioDeployer.AddService(minioCreater.GetServices())
	minioDeployer.AddDeployment(minioCreater.GetDeployment())
	minioDeployer.AddPVC(minioclaim)
	err = minioDeployer.Run()
	if err != nil {
		return err
	}

	// Consul
	consulDeployer, _ := deployer2.NewDeployer(c.kubeConfig)
	consulCreater := apps.NewConsul(spec.Namespace, spec.StorageClass)
	consulDeployer.AddService(consulCreater.GetConsulServices())
	consulDeployer.AddStatefulSet(consulCreater.GetConsulStatefulSet())
	consulDeployer.AddSecret(consulCreater.GetConsulSecrets())

	err = consulDeployer.Run()
	if err != nil {
		return err
	}

	time.Sleep(30 * time.Second)

	// Vault Init - Generate Root CA and auth certs
	// This will create the following secrets :
	// - auth-client-tls-certificate
	// - auth-server-tls-certificate
	// - vault-ca-certificate
	// - vault-tls-certificate
	// - vault-init-secret
	err = c.vaultInit(spec.Namespace)
	if err != nil {
		return err
	}

	// Vault
	vaultDeployer, _ := deployer2.NewDeployer(c.kubeConfig)

	vaultCreater := apps.NewVault(spec.Namespace, vaultConfig, map[string]string{
		"vault-tls-certificate": "/vault/tls",
	}, "/vault/tls/ca.crt")
	vaultDeployer.AddService(vaultCreater.GetVaultServices())

	// Inject auto-unseal sidecar
	vaultInit := Vault{spec.Namespace}
	vaultPod := vaultCreater.GetPod()
	vaultPod.AddVolume(components.NewSecretVolume(horizonapi.ConfigMapOrSecretVolumeConfig{
		VolumeName:      "vault-init-secret",
		MapOrSecretName: "vault-init-secret",
	}))

	vaultPod.AddContainer(vaultInit.GetSidecarUnsealContainer())

	vaultDeployer.AddDeployment(util.CreateDeployment(&horizonapi.DeploymentConfig{
		Name:      "vault",
		Namespace: spec.Namespace,
		Replicas:  util.IntToInt32(3),
	}, vaultPod))
	vaultDeployer.AddConfigMap(vaultCreater.GetVaultConfigConfigMap())
	err = vaultDeployer.Run()
	if err != nil {
		return err
	}

	time.Sleep(30 * time.Second)

	return err
}

func (c *Creater) eventStoreInit(spec *v1.RgpSpec) error {
	eventStore := NewEventstore(spec.Namespace, spec.StorageClass, 100)

	// eventstore
	eventStoreDeployer, _ := deployer2.NewDeployer(c.kubeConfig)
	eventStoreDeployer.AddStatefulSet(eventStore.GetEventStoreStatefulSet())
	eventStoreDeployer.AddService(eventStore.GetEventStoreService())

	err := eventStoreDeployer.Run()
	if err != nil {
		return err
	}

	// Create service account
	_, err = c.kubeClient.CoreV1().ServiceAccounts(spec.Namespace).Create(&v14.ServiceAccount{
		ObjectMeta: v13.ObjectMeta{
			Name: "eventstore-init",
		},
	})
	if err != nil {
		return err
	}

	// Create role
	_, err = c.kubeClient.RbacV1().Roles(spec.Namespace).Create(&v12.Role{
		ObjectMeta: v13.ObjectMeta{
			Name: "eventstore-init",
		},
		Rules: []v12.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs: []string{
					"get",
					"create",
					"update",
					"patch",
					"delete",
				},
			},
		},
	})
	if err != nil {
		return err
	}

	// Bind role to service account
	_, err = c.kubeClient.RbacV1().RoleBindings(spec.Namespace).Create(&v12.RoleBinding{
		ObjectMeta: v13.ObjectMeta{
			Name: "eventstore-init",
		},
		Subjects: []v12.Subject{
			{
				Kind: "ServiceAccount",
				Name: "eventstore-init",
			},
		},
		RoleRef: v12.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "eventstore-init",
		},
	})
	if err != nil {
		return err
	}

	time.Sleep(30 * time.Second)

	// Init Job
	err = c.startJobAndWaitUntilCompletion(spec.Namespace, 30*time.Minute, eventStore.GetInitJob())
	if err != nil {
		return err
	}

	return nil
}

// vaultInit start the vault initialization job
func (c *Creater) vaultInit(namespace string) error {
	// Init
	_, err := c.kubeClient.CoreV1().ServiceAccounts(namespace).Create(&v14.ServiceAccount{
		ObjectMeta: v13.ObjectMeta{
			Name:      "vault-init",
			Namespace: namespace,
		},
	})
	if err != nil {
		return err
	}
	_, err = c.kubeClient.RbacV1().Roles(namespace).Create(&v12.Role{
		ObjectMeta: v13.ObjectMeta{
			Name:      "vault-init",
			Namespace: namespace,
		},
		Rules: []v12.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs: []string{
					"get",
					"create",
					"update",
					"patch",
					"delete",
				},
			},
		},
	})
	if err != nil {
		return err
	}
	_, err = c.kubeClient.RbacV1().RoleBindings(namespace).Create(&v12.RoleBinding{
		ObjectMeta: v13.ObjectMeta{
			Name:      "vault-init",
			Namespace: namespace,
		},
		Subjects: []v12.Subject{
			{
				Kind: "ServiceAccount",
				Name: "vault-init",
			},
		},
		RoleRef: v12.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     "vault-init",
		},
	})
	if err != nil {
		return err
	}

	// Start job and create CM
	vaultInit := Vault{namespace}
	err = c.startJobAndWaitUntilCompletion(namespace, 30*time.Minute, vaultInit.GetJob())
	if err != nil {
		log.Print(err)
		return err
	}

	vaultInitDeploy, _ := deployer2.NewDeployer(c.kubeConfig)
	vaultInitDeploy.AddConfigMap(vaultInit.GetConfigmap())
	vaultInitDeploy.AddDeployment(vaultInit.GetDeployment())
	err = vaultInitDeploy.Run()
	if err != nil {
		log.Print(err)
		return err
	}

	return nil
}

// dbInit create the the databases
func (c *Creater) dbInit(namespace string, pw string) error {
	databaseName := "postgres"
	hostName := fmt.Sprintf("postgres.%s.svc.cluster.local", namespace)

	postgresDB, err := OpenDatabaseConnection(hostName, databaseName, "postgres", pw, "postgres")
	// log.Infof("Db: %+v, error: %+v", db, err)
	if err != nil {
		return fmt.Errorf("unable to open database connection for %s database in the host %s due to %+v", databaseName, hostName, err)
	}

	for {
		log.Debug("executing SELECT 1")
		_, err := postgresDB.Exec( "SELECT 1;")
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	_, err = postgresDB.Exec("CREATE DATABASE \"tools-portfolio\";")
	if err != nil {
		return err
	}
	_, err = postgresDB.Exec("CREATE DATABASE \"rp-portfolio\";")
	if err != nil {
		return err
	}
	_, err = postgresDB.Exec("CREATE DATABASE \"report-service\";")
	if err != nil {
		return err
	}
	_, err = postgresDB.Exec("CREATE DATABASE \"issue-manager\";")
	if err != nil {
		return err
	}
	return nil
}

// OpenDatabaseConnection open a connection to the database
func OpenDatabaseConnection(hostName string, dbName string, user string, password string, sqlType string) (*sql.DB, error) {
	// Note that sslmode=disable is required it does not mean that the connection
	// is unencrypted. All connections via the proxy are completely encrypted.
	log.Debug("attempting to open database connection")
	dsn := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable connect_timeout=10", hostName, dbName, user, password)
	db, err := sql.Open(sqlType, dsn)
	//defer db.Close()
	if err == nil {
		log.Debug("connected to database ")
	}
	return db, err
}

func (c *Creater) startJobAndWaitUntilCompletion(namespace string, timeoutValue time.Duration, job *v1_batch.Job) error {
	job, err := c.kubeClient.BatchV1().Jobs(namespace).Create(job)
	if err != nil {
		return err
	}
	timeout := time.After(timeoutValue)
	tick := time.NewTicker(10 * time.Second)

L:
	for {
		select {
		case <-timeout:
			tick.Stop()
			return fmt.Errorf("job failed")

		case <-tick.C:
			job, err = c.kubeClient.BatchV1().Jobs(job.Namespace).Get(job.Name, v13.GetOptions{})
			if err != nil {
				tick.Stop()
				return err
			}
			if job.Status.Succeeded > 0 {
				tick.Stop()
				break L
			}
		}
	}
	return nil
}
