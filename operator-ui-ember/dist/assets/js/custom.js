$('#prerequisites').click(function(){
    $.ajax({
        url: '/documentation/prerequisites',
        type: 'GET',
        success: function (response) {
            $("#main").html(response).end().appendTo($('#main'));
        },
        error: function (error) {
            swal({
                title: "Failure",
                text: "Unable to prerequisites documentation from server.",
                type: 'error'
            });
        }
    });
});

$('#install-operator').click(function(){
    $.ajax({
        url: '/documentation/install-operator',
        type: 'GET',
        success: function (response) {
            $("#main").html(response).end().appendTo($('#main'));
        },
        error: function (error) {
            swal({
                title: "Failure",
                text: "Unable to operation installation documentation from server.",
                type: 'error'
            });
        }
    });
});

$('#deploy-on-prem').click(function(){
    $.ajax({
        url: '/documentation/onprem',
        type: 'GET',
        success: function (response) {
            $("#main").html(response).end().appendTo($('#main'));
        },
        error: function (error) {
            swal({
                title: "Failure",
                text: "Unable to on-prem installation documentation from server.",
                type: 'error'
            });
        }
    });
});

$('#deploy-on-gcp').click(function(){
    $.ajax({
        url: '/documentation/gcp',
        type: 'GET',
        success: function (response) {
            $("#main").html(response).end().appendTo($('#main'));
        },
        error: function (error) {
            swal({
                title: "Failure",
                text: "Unable to gcp documentation from server.",
                type: 'error'
            });
        }
    });
});

$('#deploy-on-aws').click(function(){
    $.ajax({
        url: '/documentation/aws',
        type: 'GET',
        success: function (response) {
            $("#main").html(response).end().appendTo($('#main'));
        },
        error: function (error) {
            swal({
                title: "Failure",
                text: "Unable to aws documentation from server.",
                type: 'error'
            });
        }
    });
});

$('#deploy-on-azure').click(function(){
    $.ajax({
        url: '/documentation/azure',
        type: 'GET',
        success: function (response) {
            $("#main").html(response).end().appendTo($('#main'));
        },
        error: function (error) {
            swal({
                title: "Failure",
                text: "Unable to azure documentation from server.",
                type: 'error'
            });
        }
    });
});