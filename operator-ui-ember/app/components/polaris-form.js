import Component from '@ember/component';

export default Component.extend({
    polarisConfig: {
        SMTPHost: "",
        SMTPPort: "",
        SMTPUsername: "",
        SMTPPassword: "",
        environmentName: "",
        environmentAddress: "",
        namespace: "",
        storageClass: "",
        internalPostgresInstance: false,
        externalPostgresInstance: false,
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
        },
        setInternalPostgresInstance() {
            this.polarisConfig.internalPostgresInstance = true
            this.polarisConfig.externalPostgresInstance = false
        },
        setExternalPostgresInstance() {
            this.polarisConfig.internalPostgresInstance = false
            this.polarisConfig.externalPostgresInstance = true
        }
    }
});
