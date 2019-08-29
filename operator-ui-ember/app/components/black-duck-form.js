import Component from '@ember/component';

export default Component.extend({
    deployed: false,
    actions: {
        deployBlackDuck() {
            alert("Running AJAX for Black Duck...")
            //alert(model.name)
            var dataString = "Hello Black Duck";
            $.ajax({
                type: "POST",
                url: "http://localhost:8081/",
                data: dataString,
                success: function () {
                    alert("success")
                }
            });
        }
    }
});
