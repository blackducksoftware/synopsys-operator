import Component from '@ember/component';

export default Component.extend({
    spec: {
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
            alert("Running AJAX for Polaris...")
            $.ajax({
                type: "POST",
                url: "http://localhost:8081/",
                data: JSON.stringify(this.spec),
                success: function () {
                    alert("success")
                }
            });
        },
        setInternalPostgresInstance() {
            this.spec.internalPostgresInstance = true
            this.spec.externalPostgresInstance = false
        },
        setExternalPostgresInstance() {
            this.spec.internalPostgresInstance = false
            this.spec.externalPostgresInstance = true
        }
    }
});
