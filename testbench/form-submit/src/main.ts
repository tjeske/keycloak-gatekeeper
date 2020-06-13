require('../semantic/dist/components/api')
require('../semantic/dist/components/form')
import '../semantic/dist/semantic.min.css'
import * as $ from 'jquery'
import 'form-serializer'

// $(document).ready(function() {
// $('.ui.button').on('click', function() {
//     $('form').submit();
// });
// $.fn.api.settings.api = {
//     'add user'      : '/udesk/updateTemplate',
// };
// $('form').api({
//     url: 'http://localhost:3000/udesk/updateTemplate',
//     method: 'post',
//     serializeForm : true,
//     data          : {
//         specifiedKey: 'some data' //Not sent
//     },
//     beforeSend : function (settings) {
//         settings.data.beforeKey = 'other data'; //Not sent
//         return settings;
//     },
//     response: function(settings) {
//         console.log(settings);

//     //   alert(settings.data.specifiedKey);
//     //   alert(settings.data.beforeKey);
//     //   alert(settings.data.formKey);
//     },
//     onRequest: function (result, a) {
//         console.log(a);

//     }
// });
// // $('.ui.form .submit.button')
// // .api({
// //     url: '/udesk/updateTemplatesdfsd',
// //     method : 'post',
// //     serializeForm: true,
// //     beforeSend: function(settings) {
// //     },
// //     onSuccess: function(data) {
// //     }
// // });
// })

// geht nicht!!!!
$(document).ready(function () {

    $('.ui.button').on('click', function () {
        $('form').submit();
    });
    $('.form.button').form(
        {
            on: 'blur',
            fields: {
                username: {
                    identifier: 'username',
                    rules: [{
                        type: 'empty',
                        prompt: 'Username cannot be empty'
                    }]
                },
                password: {
                    identifier: 'password',
                    rules: [{
                        type: 'empty',
                        prompt: 'Password cannot be emtpy'
                    }]
                }
            },
            onSuccess: function (event) {
                $('#formresult').hide();
                $('#formresult').text('');
                event.preventDefault();
                return false;
            }

        }
    )
        .api({
            url: "/udesk/updateTemplate/",
            method: 'post',
            serializeForm: true,
            data: { a: "JHKHKH" }, //new FormData(document.querySelector('form')),
            onSuccess: function (result) {
                $('#formresult').show();
                console.log("SUCCESS");

                return false;
            }
        });


});