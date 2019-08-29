import Route from '@ember/routing/route';

export default Route.extend({
    actions: {
        deployPolaris() {
            alert("Running AJAX for Polaris...")
            //alert(model.name)
            var dataString = "Hello Polaris";
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
