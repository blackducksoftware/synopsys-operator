$(function() {
    $('#polaris-options-form').on("submit",function(e) {
        e.preventDefault(); // cancel the actual submit
        swal({
            title: "Confirmation",
            text: "Are you sure you want to proceed?",
            type: "info",
            showCancelButton: true,
            closeOnConfirm: false,
            showLoaderOnConfirm: true
          }, function () {
            $.ajax({
                url: '/install',
                type: 'POST',
                contentType: 'application/x-www-form-urlencoded',
                data: $('#polaris-options-form').serialize(),
                success: function (response) {
                    swal({
                        title: "Success",
                        text: response.responseText,
                        type: 'success'
                    },
                    function(){
                        location.href = '/complete';
                    });
                },
                error: function (error) {
                    swal({
                        title: "Failure",
                        text: error.responseText,
                        type: 'error'
                    });
                }
            });
        });          
    });
});

$(function() {
    $('#operator-deploy-form').on("submit",function(e) {
        e.preventDefault(); // cancel the actual submit
        swal({
            title: "Confirmation",
            text: "Are you sure you want to proceed?",
            type: "info",
            showCancelButton: true,
            closeOnConfirm: false,
            showLoaderOnConfirm: true
          }, function () {
            $.ajax({
                url: '/deploy',
                type: 'POST',
                contentType: 'application/x-www-form-urlencoded',
                data: $('#operator-deploy-form').serialize(),
                success: function (response) {
                    swal({
                        title: "Success",
                        text: response.responseText,
                        type: 'success'
                    },
                    function(){
                        location.href = '/operator';
                    });
                },
                error: function (error) {
                    swal({
                        title: "Failure",
                        text: error.responseText,
                        type: 'error'
                    });
                }
            });
        });          
    });
});

$(document).on('change', 'input[type=radio][name=postgresInstanceType]', function() {
    if (this.value == 'external') {
        $.ajax({
            url: '/postgres_form',
            type: 'GET',
            success: function (response) {
                $("#postgres").html(response).end().appendTo($('#postgres'));
            },
            error: function (error) {
                swal({
                    title: "Failure",
                    text: "Unable to fetch postgres form from server.",
                    type: 'error'
                });
            }
        });
    }
    else if (this.value == 'internal') {
        $("#postgres").empty();
    }
});

$(document).on('change', 'input[type=radio][name=sslCertsType]', function() {
    if (this.value == 'custom') {
        $.ajax({
            url: '/ssl_form',
            type: 'GET',
            success: function (response) {
                $("#ssl").html(response).end().appendTo($('#ssl'));
            },
            error: function (error) {
                swal({
                    title: "Failure",
                    text: "Unable to fetch ssl form from server.",
                    type: 'error'
                });
            }
        });
    }
    else if (this.value == 'self-signed') {
        $("#ssl").empty();
    }
});

$(document).on('change', 'input[type=radio][name=deploymentPlatform]', function() {
    if (this.value == 'on-prem') {
        $.ajax({
            url: '/on_prem_deploy_form',
            type: 'GET',
            success: function (response) {
                $("#deployOptions").html(response).end().appendTo($('#deployOptions'));
            },
            error: function (error) {
                swal({
                    title: "Failure",
                    text: "Unable to fetch on-prem deploy form from server.",
                    type: 'error'
                });
            }
        });
    }
    else if (this.value == 'aws') {
        $.ajax({
            url: '/aws_deploy_form',
            type: 'GET',
            success: function (response) {
                $("#deployOptions").html(response).end().appendTo($('#deployOptions'));
            },
            error: function (error) {
                swal({
                    title: "Failure",
                    text: "Unable to fetch aws deploy form from server.",
                    type: 'error'
                });
            }
        });
    }
    else if (this.value == 'gcp') {
        $.ajax({
            url: '/gcp_deploy_form',
            type: 'GET',
            success: function (response) {
                $("#deployOptions").html(response).end().appendTo($('#deployOptions'));
            },
            error: function (error) {
                swal({
                    title: "Failure",
                    text: "Unable to fetch gcp deploy form from server.",
                    type: 'error'
                });
            }
        });
    }
    else if (this.value == 'azure') {
        $.ajax({
            url: '/azure_deploy_form',
            type: 'GET',
            success: function (response) {
                $("#deployOptions").html(response).end().appendTo($('#deployOptions'));
            },
            error: function (error) {
                swal({
                    title: "Failure",
                    text: "Unable to fetch gcp deploy form from server.",
                    type: 'error'
                });
            }
        });
    }
});

$(document).on('click', 'button[id=smtp-button]', function() {
    swal({
        html: true,
        title: "Test SMTP Settings",
        type: "info",
        showCancelButton: true,
        closeOnConfirm: false,
        showLoaderOnConfirm: true,
        text: `<div class='row'>
                <div class='col-6'>
                    <label for='smtpSender'>Sender Email</label>
                    <input type='text' class='form-control' name='smtpSender'
                        id='smtpSender' placeholder='Please enter sender email' />
                </div>
                <div class='col-6'>
                    <label for='smtpReceiver'>Receiver Email</label>
                    <input type='text' class='form-control' name='smtpReceiver'
                        id='smtpReceiver' placeholder='Please enter receiver email' />
                </div>
            </div>`
    }, function () {
        $.ajax({
            url: '/test_smtp_settings',
            type: 'POST',
            contentType: 'application/x-www-form-urlencoded',
            data: {
                "smtpHost": $('#smtpHost').val(),
                "smtpPort": $('#smtpPort').val(),
                "smtpUsername": $('#smtpUsername').val(),
                "smtpPassword": $('#smtpPassword').val(),
                "smtpSender": $('#smtpSender').val(),
                "smtpReceiver": $('#smtpReceiver').val()
            },
            success: function (response) {
                swal({
                    title: "SMTP settings are Valid",
                    text: response.responseText,
                    type: 'success'
                });
            },
            error: function (error) {
                swal({
                    title: "SMTP settings are invalid",
                    text: error.responseText,
                    type: 'error'
                });
            }
        });
    });
});

$(document).on('click', 'a[id=deploy_polaris]', function() {
    $.ajax({
        url: '/operator/status',
        type: 'GET',
        success: function (response) {
            if (response.isInstalled){
                location.href = '/deploy_polaris';
            }else{
                swal({
                    title: "Warning",
                    text: "Polaris Operator is not deployed. Redirecting to operator installtion page. !",
                    type: 'info'
                },
                function(){
                    location.href = '/operator';
                });
            }
        },
        error: function (error) {
            swal({
                title: "Failure",
                text: "Unable to operator status from server.",
                type: 'error'
            });
        }
    });
});