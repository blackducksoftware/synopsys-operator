import Component from '@ember/component';

export default Component.extend({
    blackDuckConfig: {
        name: "",
        namespace: "",
        version: "",
        licenseKey: "",
        dbMigrate: false,
        size: "small",
        exposeService: "",
        blackDuckType: "",
        useBinaryUploads: false,
        enableSourceUploads: false,
        livenessProbes: false,
        persistentStorage: true,
        cloneDB: "",
        PVCStorageClass: "",
        scanType: "",
        externalDatabase: false,
        postgresSQLUserPassword: "",
        postgresSQLAdminPassword: "",
        postgresSQLPostgresPassword: "",
        certificateName: "",
        customCACertificateAuthentication: false,
        proxyRootCertificate: "",
    },
    actions: {
        deployBlackDuck() {
            alert("Running AJAX for Black Duck...")
            $.ajax({
                type: "POST",
                url: "http://localhost:8081/api/deploy_black_duck",
                data: JSON.stringify(this.blackDuckConfig),
                success: function () {
                    alert("success")
                }
            });
        }
    }
});
