import Component from '@ember/component';

export default Component.extend({
    deployed: false,
    actions: {
        deploy() {
            alert("Running AJAX...")
            var dataString = "hello there";
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
