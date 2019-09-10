import Component from '@ember/component';

export default Component.extend({
    polarisConfig: {
        version: "",
        environmentName: "",
        environmentDNS: "",
        storageClass: "",
        namespace: "",
        name: "",
        imagePullSecrets: "",
        postgresHost: "",
        postgresPort: "",
        postgresUsername: "",
        postgresPassword: "",
        postgresSize: "",
        smtpHost: "",
        smtpPort: "",
        smtpUsername: "",
        smtpPassword: "",
        uploadServerSize: "",
        eventstoreSize: ""
    },
    actions: {
        deployPolaris() {
            $.ajax({
                type: "POST",
                url: "/api/deploy_polaris",
                data: JSON.stringify(this.polarisConfig),
                success: function () {
                    alert("success")
                }
            });
        }
    }
});
