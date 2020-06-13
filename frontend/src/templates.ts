import { addApp, initTemplateDropDown, Template } from 'shared'
import 'form-serializer'


type FormData2 = {
    name: string,
    internalPort: number,
    access: AccessForm[]
}

type AccessForm = {
    user: string,
    run: boolean,
    control: boolean,
    modify: boolean
}

function loadTemplate(name: string): void {

    // get templates
    $.ajax({
        url: "/udesk/getTemplate/" + name,
        success: function (result: { data: Template }) {
            let element = result.data
            if (element == undefined) {
                ($('body') as any).toast({
                    class: 'error',
                    message: `Cannot load template!`
                });
                return
            }
            addApp(element.name)
            if ('name' in element) {
                $('#name').val(element.name)
            }

            if ('internalPort' in element) {
                $('#internalPort').val(element.internalPort)
            }

            $("#params-table-body").empty()
            $("#params-table-body").html("<tr></tr>");
            if ('params' in element) {
                if (element.params != null) {
                    for (let [key, value] of Object.entries(element.params)) {
                        addParameter(key, value, "#params-table-body");
                    }
                }
            }
            if ('dockerfile' in element) {
                $('#dockerfile-textarea').val(element.dockerfile)
            }

            $("#files-table-body").empty()
            $("#files-table-body").html(`<tr class="entry"></tr>`);
            if ('files' in element) {
                if (element.files != null) {
                    for (let [name, content] of Object.entries(element.files)) {
                        addFile(name, content);
                    }
                }
            }

            $("#access-table-body").children().remove()
            $("#access-table-body").html(`<tr class="entry"></tr>`);
            if ('access' in element) {
                if (element.access != null) {
                    for (let userName in element.access) {
                        if ('permissions' in element.access[userName]) {
                            let permissions = element.access[userName].permissions.split(",").map((it: string) => it.trim())
                            addAccess(userName, permissions.includes("readable"), permissions.includes("controlable"), permissions.includes("modifiable"))
                        }
                    }
                }
            }

            updateForm()
        },
        error: function (result) {
            ($('body') as any).toast({
                class: 'error',
                message: `Cannot get template list from server`
            });
        }
    });
}

export function addParameter(paramName: string, defaultValue: string, paramsTableBodyId: string) {
    let row = `
    <tr>
        <td>
            <div class="ui input"><input name="paramName" type="text" value="${paramName}">
            </div>
        </td>
        <td>=</td>
        <td>
            <div class="ui input"><input name="paramValue" type="text" value="${defaultValue}">
            </div>
        </td>
        <td>
            <div class="ui small icon">
                <i class="ui right red icon delete"></i>
            </div>
        </td>
    </tr>`

    let paramRow = $(row);
    paramRow.hide();
    $(`${paramsTableBodyId} tr:last-child`).after(paramRow);
    $(`${paramsTableBodyId} tr:last-child td:last-child div`).on('click', function () {
        $(this).parent().parent().fadeOut(300, function () { $(this).remove(); });
    });
    paramRow.fadeIn("slow");
}

export function addFile(fileName: string, fileContent: string) {
    let row = `
    <tr class="entry">
        <td>
            <table class="ui very basic table" style="width=100%;">
                <tbody>
                    <tr>
                        <td style="width: 100%;">
                            <div class="ui black input"><input name="fileName" type="text" value="${fileName}">
                            </div>
                        </td>
                        <td>
                            <div class="ui small icon">
                                <i class="ui right red icon delete remove-file-btn"></i>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
            <textarea name="fileContent" style="font-family:monospace;">${fileContent}</textarea>
        </td>
    </tr>`

    let paramRow = $(row);
    paramRow.hide();
    $('#files-table-body .entry:last').after(paramRow);
    $('#files-table-body .entry:last .remove-file-btn').on('click', function () {
        $(this).parent().parent().parent().parent().parent().parent().fadeOut(300, function () { $(this).remove(); });
    });
    paramRow.fadeIn("slow");
}

export function addAccess(userName: string, isReadable: boolean, isControlable: boolean, isModifiable: boolean) {
    let row = `
        <tr>
            <td>
                <div class="inline fields">
                    <label><i class="user icon"></i></label>
                    <div class="ui search">
                        <input class="prompt" type="text" placeholder="... etc" value="${userName}">
                        <div class="results"></div>
                    </div>
                    <label style="width: auto; margin-left: auto;"><i class="right icon delete ui red remove-access-btn"></i></label>
                </div>
            </td>
            <td class="ui center aligned">
                <div class="ui checkbox">
                    <input type="checkbox" class="AccessModifierRun" ${isReadable ? ` checked` : ``}>
                </div>
            </td>
            <td class="ui center aligned">
                <div class="ui checkbox">
                    <input type="checkbox" class="AccessModifierControl" ${isControlable ? ` checked` : ``}>
                </div>
            </td>
            <td class="ui center aligned">
                <div class="ui checkbox">
                    <input type="checkbox" class="AccessModifierModify" ${isModifiable ? ` checked` : ``}>
                </div>
            </td>
        </tr>`

    let paramRow = $(row);
    paramRow.hide();
    $('#access-table-body tr:last').before(paramRow);
    $('#access-table-body tr:nth-last-child(2) .checkbox').checkbox();
    $('#access-table-body tr:nth-last-child(2) .remove-access-btn').on('click', function () {
        $(this).parent().parent().parent().parent().fadeOut(300, function () { $(this).remove(); });
    });
    $('.ui.search').search({
        debug: true,
        apiSettings: {
            action: 'search',
            url: '/udesk/searchUser/{query}'
        },
    });
    paramRow.fadeIn("slow");
}

function initAppDropDown() {
    // make checkboxes visible without label
    $('.checkbox').checkbox();
    $('#app-dropdown').dropdown({
        onChange: function (value, text, $selectedItem) {
            $('#app-dropdown').dropdown('set selected', value);
            if (value != "") {
                loadTemplate(value)
            }
        }
    });
}

function updateForm() {
    // stop the form from submitting normally 
    // $('.ui.form').submit(function (e) {
    //     //e.preventDefault(); usually use this, but below works best here.
    //     return false;
    // });
    // $('.ui.submit').on('click', function() {
    //     console.log("KJH");

    //     $('form').submit();
    // });
    // $('form')
    //     .api({
    //         on: 'change',
    //         method: 'post',
    //         url: '/udesk/updateTemplate',
    //         data: {
    //             a: "HGHJGJ"
    //         },
    //         serializeForm: true,
    //         beforeSend: function (settings) {

    //             settings.data.access = []
    //             $("#access-table tbody tr").each(function() {
    //                 if ($(this).has("td").length) {
    //                     //console.log($(this).find('input[name="AccessName"]').attr('value'));
    //                     let user = $(this).find('input.prompt').val()
    //                     let run = $(this).find('input.AccessModifierRun:checkbox').prop('checked')
    //                     let control = $(this).find('input.AccessModifierControl:checkbox').prop('checked')
    //                     let modify = $(this).find('input.AccessModifierModify:checkbox').prop('checked')
    //                     settings.data.access.push({
    //                         user: user,
    //                         run: run,
    //                         control: control,
    //                         modify: modify
    //                     })
    //                 }
    //             })

    //             console.log(settings)
    //             settings.data.username = 'New User';

    //             // form data is editable in before send
    //             if (settings.data.username == '') {
    //                 settings.data.username = 'New User';
    //             }
    //             // open console to inspect object
    //             console.log(settings.data);

    //             return settings;
    //         },
    //         onResponse: function(settings: any) {
    //             alert(settings.data.specifiedKey);
    //             alert(settings.data.beforeKey);
    //             alert(settings.data.formKey);
    //           }
    //     });


    $('.ui.form').submit(function (e) {
        //e.preventDefault(); usually use this, but below works best here.
        return false;
    });
    $('.ui.form')
        .form({
            onSuccess: function () {

                let access2: AccessForm[] = []
                $("#access-table tbody tr").each(function () {
                    if ($(this).has("td").length) {
                        //console.log($(this).find('input[name="AccessName"]').attr('value'));
                        let user = $(this).find('input.prompt').val()
                        let run = $(this).find('input.AccessModifierRun:checkbox').prop('checked')
                        let control = $(this).find('input.AccessModifierControl:checkbox').prop('checked')
                        let modify = $(this).find('input.AccessModifierModify:checkbox').prop('checked')
                        access2.push({
                            user: user as string,
                            run: run as boolean,
                            control: control as boolean,
                            modify: modify as boolean
                        })
                    }
                })

                let formData: FormData2 = {
                    name: getFieldValue("name") as string,
                    internalPort: getFieldValue("internalPort") as number,
                    access: access2
                };

                console.log(formData);

                let uuid = 123
                $.ajax({
                    url: "/udesk/updateTemplate/" + uuid,
                    type: 'POST',
                    data: [
                        {
                            "access": "ABC"
                        },
                        {
                            "access": "DEF"
                        }
                    ],
                    success: function (result: { data: Template }) {
                        ($('body') as any).toast({
                            class: 'success',
                            message: `Template updated successfully!`
                        });
                    },
                    error: function (result) {
                        ($('body') as any).toast({
                            class: 'error',
                            message: `Cannot update template`
                        });
                    }
                });
            },
            fields: {
                name: {
                    identifier: 'name',
                    rules: [
                        {
                            type: 'empty',
                            prompt: 'Name should not be empty'
                        }
                    ]
                },
                internalPort: {
                    identifier: 'internalPort',
                    rules: [
                        {
                            type: 'decimal',
                            prompt: 'Port should be decimal'
                        }
                    ]
                },
                paramName: {
                    identifier: 'paramName',
                    rules: [
                        {
                            type: 'empty',
                            prompt: 'Parameter name cannot be empty'
                        }
                    ]
                },
                paramValue: {
                    identifier: 'paramValue',
                    rules: [
                        {
                            type: 'empty',
                            prompt: 'Parameter value cannot be empty'
                        }
                    ]
                },
                dockerfile: {
                    identifier: 'dockerfile',
                    rules: [
                        {
                            type: 'empty',
                            prompt: 'Dockerfile cannot be empty'
                        }
                    ]
                },
                fileName: {
                    identifier: 'paramValue',
                    rules: [
                        {
                            type: 'empty',
                            prompt: 'File name cannot be empty'
                        }
                    ]
                },
                fileContent: {
                    identifier: 'fileContent',
                    rules: [
                        {
                            type: 'empty',
                            prompt: 'File content cannot be empty'
                        }
                    ]
                },
                access: {
                    identifier: 'fileContent',
                    rules: [
                        {
                            type: 'empty',
                            prompt: 'File content cannot be empty'
                        }
                    ]
                }
            }
        });
}

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

function getFieldValue(fieldId: string) {
    // 'get field' is part of Semantics form behavior API
    return $('.ui.form').form('get field', fieldId).val();
}

