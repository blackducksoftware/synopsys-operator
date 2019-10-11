import Component from '@ember/component';
import $ from 'jquery';

export default Component.extend({
    blackDuckConfig: {
        name: "",
        namespace: "",
        version: "2019.6.2",
        licenseKey: "",
        dbMigrate: false,
        size: "small",
        exposeService: "None",
        blackDuckType: "",
        useBinaryUploads: false,
        enableSourceUploads: false,
        livenessProbes: false,
        persistentStorage: true,
        cloneDB: "",
        PVCStorageClass: "",
        scanType: "",
        externalDatabase: false,
        externalPostgresSQLHost: "",
        externalPostgresSQLPort: "",
        externalPostgresSQLAdminUser: "",
        externalPostgresSQLAdminPassword: "",
        externalPostgresSQLUser: "",
        externalPostgresSQLUserPassword: "",
        enableSSL: false,
        postgresSQLUserPassword: "",
        postgresSQLAdminPassword: "",
        postgresSQLPostgresPassword: "",
        certificateName: "",
        customCACertificateAuthentication: false,
        proxyRootCertificate: "",
        containerImageTags: "",
        environmentVariables: "USE_BINARY_UPLOADS:0\nBROKER_USE_SSL:yes\nHTTPS_VERIFY_CERTS:yes\nRABBITMQ_SSL_FAIL_IF_NO_PEER_CERT:false\nDATA_RETENTION_IN_DAYS:180\nENABLE_SOURCE_UPLOADS:false\nMAX_TOTAL_SOURCE_SIZE_MB:4000\nIPV4_ONLY:0\nUSE_ALERT:0\nRABBIT_MQ_PORT:5671\nSCANNER_CONCURRENCY:1\nRABBITMQ_DEFAULT_VHOST:protecodesc",
        nodeAffinityJSON: ""
    },
    actions: {
        deployBlackDuck() {
            $.ajax({
                type: "POST",
                url: "/api/deploy_black_duck",
                data: JSON.stringify(this.blackDuckConfig),
                success: function () {
                    alert("success")
                }
            });
        }
    }
});
