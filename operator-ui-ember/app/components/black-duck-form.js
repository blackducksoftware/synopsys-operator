import Component from '@ember/component';

export default Component.extend({
    name: "",
    namespace: "",
    version: "",
    actions: {
        deployBlackDuck() {
            alert("Running AJAX for Black Duck...")
            //alert(model.name)
            var BlackDuckSpecData = {
                name: this.name,
                namespace: this.namespace,
                version: this.version
            }
            alert(BlackDuckSpecData)
            var dataString = "Hello Black Duck";
            $.ajax({
                type: "POST",
                url: "http://localhost:8081/",
                data: JSON.stringify(BlackDuckSpecData),
                success: function () {
                    alert("success")
                }
            });
        }
    }
});
