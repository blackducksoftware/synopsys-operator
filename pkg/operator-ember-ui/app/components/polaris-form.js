import Component from '@ember/component';
import $ from 'jquery';

export default Component.extend({
    receivedValidPolarisInstance: false,
    instanceNamespace: "",
    polarisConfig: {
        version: "",
        environmentDNS: "",
        storageClass: "",
        namespace: "",
        googleServiceAccountPath: "",
        //postgresHost: "",
        //postgresPort: "",
        postgresUsername: "",
        postgresPassword: "",
        postgresSize: "",
        smtpHost: "",
        smtpPort: "",
        smtpUsername: "",
        smtpPassword: "",
        smtpSenderEmail: "",
        uploadServerSize: "",
        eventstoreSize: "",
        mongoDBSize: "",
        downloadServerSize: "",
        enableReporting: false,
        reportStorageSize: "",
        organizationDescription: "",
        organizationName: "",
        organizationAdminName: "",
        organizationAdminUsername: "",
        organizationAdminEmail: "",
        coverityLicensePath: "",
    },
    actions: {
        deployPolaris() {
            $.ajax({
                type: "POST",
                url: "/api/ensure_polaris",
                data: JSON.stringify(this.polarisConfig),
                success: function () {
                    alert("success - sent request to server")
                }
            });
        },
        getPolaris() {
            $.ajax({
                type: "POST",
                url: "/api/get_polaris",
                data: this.instanceNamespace,
            }).then((ajaxResults) => this.send("populatePolarisConfig", ajaxResults))
        },
        populatePolarisConfig(config) {
            let configJson = JSON.parse(config)
            this.set("receivedValidPolarisInstance", true)
            this.set('polarisConfig', configJson)
        }
    }

});
