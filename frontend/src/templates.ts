import { addAccess, addApp, addFile, addParameter, initAppDropDown, initTemplateDropDown } from 'shared'

$(document).ready(function () {
    initAppDropDown()

    $('#app-edit-dropdown').dropdown();
    $('#add-app-btn').click(function () {
        $('.ui.modal').modal({
            onApprove: function (e) {
                addApp($('#new-app-name').val() as string)
            },
        }).modal('show')
    });
    $('#remove-app-btn').click(function () {
        $('#app-remove-modal-name').html("jhgjhg")
        $('#app-remove-modal').modal({
            onApprove: function (e) {
                // apps.forEach(element => {
                //     element.selected = false;
                // });

                // $('.ui.dropdown').dropdown('change values', apps)
            },
        }).modal('show')
    });
    $('#add-new-parameter-btn').click(function () {
        addParameter("", "", "#params-table-body")
    })

    // $('#file-dropdown').dropdown({
    //     values: config[0].files.map(function (x) {
    //         return { name: x.name }
    //     })
    // });

    $('#add-new-file-btn').click(function () {
        addFile("", "")
    })

    $('#add-new-access-btn').click(function () {
        addAccess("", false, false, false)
    })

    initTemplateDropDown()
});