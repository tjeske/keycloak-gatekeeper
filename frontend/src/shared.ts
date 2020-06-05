require('../semantic/dist/components/api')
require('../semantic/dist/components/checkbox')
require('../semantic/dist/components/dimmer')
require('../semantic/dist/components/dropdown')
require('../semantic/dist/components/modal')
require('../semantic/dist/components/search')
require('../semantic/dist/components/toast')
require('../semantic/dist/components/transition')
import '../semantic/dist/semantic.min.css'
import 'datatables.net/js/jquery.dataTables'
import 'datatables.net-se/js/dataTables.semanticui'
import 'datatables.net-se/css/dataTables.semanticui.css'
import * as $ from 'jquery'
import 'universedesk.css'

interface Access {
    permissions: string
    args: Map<string, string>
}
export interface Template {
    name: string
    params: Map<string, string>
    dockerfile: string
    files: Map<string, string>
    access: { [key: string]: Access; }
    internalPort: string
}

let config: Template[] = []

/**
 * Initialize drop down list for app templates.
 */
export function initTemplateDropDown() {
    // get templates
    $.ajax({
        url: "/udesk/getTemplates",
        success: function (result: {data : Template[]}) {
            config = result.data
            let values = result.data.map(it => {
                let values: { [key: string]: string; } = {}
                values['name'] = it.name
                values['value'] = it.name
                return values
            });
            ($('#app-dropdown') as any).dropdown('change values', values);
            if (config.length > 0) {
                $('#app-dropdown').dropdown('set selected', config[0].name);
            }
        },
        error: function (result) {
            ($('body') as any).toast({
                class: 'error',
                message: `Cannot get template list from server`
            });
        }
    });
}

function loadApp(name: string) {
    let element = config.find(it => 'name' in it && it.name == name)
    if (element != undefined) {
        addApp(element.name)
        if ('internalPort' in element) {
            addInternalPort(element.internalPort)
        }
        if ('params' in element) {
            if (element.params != null) {
                for (let [key, value] of Object.entries(element.params)) {
                    addParameter(key, value, "#params-table-body");
                }
            }
        }
        if ('dockerfile' in element) {
            addDockerfile(element.dockerfile)
        }
        if ('files' in element) {
            if (element.files != null) {
                for (let [name, content] of Object.entries(element.files)) {
                    addFile(name, content);
                }
            }
        }
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
    }
}

export function addApp(appName: string) {
    // apps.forEach(element => {
    //     element.selected = false;
    // });
    // apps = apps.concat({
    //     name: appName,
    //     value: apps.length + 1,
    //     selected: true
    // });
    // $('.ui.dropdown').dropdown('change values', apps)
}

function addInternalPort(internalPort: string) {
    $('#internalPort').val(internalPort)
}

export function addParameter(paramName:string, defaultValue: string, paramsTableBodyId:string) {
    let row = `
    <tr>
        <td>
            <div class="ui black input"><input type="text" value="${paramName}">
            </div>
        </td>
        <td>=</td>
        <td>
            <div class="ui black input"><input type="text" value="${defaultValue}"></div>
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

function addDockerfile(fileContent: string) {
    $('#dockerfile-textarea').val("jhgjhg")
}

export function addFile(fileName: string, fileContent: string) {
    let row = `
    <tr class="entry">
        <td>
            <table class="ui very basic table" style="width=100%;">
                <tbody>
                    <tr>
                        <td style="width: 100%;">
                            <div class="ui black input"><input type="text" value="${fileName}">
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
            <textarea name="">${fileContent}</textarea>
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
                <div class="ui checkbox"><input type="checkbox" name="example"${isReadable ? ` checked` : ``}></div>
            </td>
            <td class="ui center aligned">
                <div class="ui checkbox"><input type="checkbox" name="example"${isControlable ? ` checked` : ``}></div>
            </td>
            <td class="ui center aligned">
                <div class="ui checkbox"><input type="checkbox" name="example"${isModifiable ? ` checked` : ``}></div>
            </td>
        </tr>`

    let paramRow = $(row);
    paramRow.hide();
    $('#access-table-body tr:last-child').after(paramRow);
    $('#access-table-body tr:last-child .checkbox').checkbox();
    $('#access-table-body tr:last-child .remove-access-btn').on('click', function () {
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

export function showRemoveContainerModal(containerName: string, containerId: string, table: DataTables.Api) {
    let modal = `
        <div class="ui tiny modal" id="container-remove-modal">
            <div class="content">
                <p>Do you really want to remove container <strong>${containerName}</strong>?</p>
            </div>
            <div class="actions">
            <div class="ui negative button">
              No
            </div>
            <div class="ui positive right labeled icon button">
              Yes
              <i class="checkmark icon"></i>
            </div>
          </div>
        </div>`
    $(modal).modal({
        onApprove: function (e) {
            $.ajax({
                url: "/udesk/removeApp/" + containerId,
                success: function (result) {
                    ($('body') as any).toast({
                        class: 'success',
                        message: `Removed container successfully!`
                    });
                    table.ajax.reload();
                },
                error: function (result) {
                    ($('body') as any).toast({
                        class: 'error',
                        message: `Removing container failed!`
                    });
                }
            });
        },
    }).modal('show')
}

export function initAppDropDown() {
    // make checkboxes visible without label
    $('.checkbox').checkbox();
    $('#app-dropdown').dropdown({
        onChange: function (value, text, $selectedItem) {
            $('#app-dropdown').dropdown('set selected', value);
            loadApp(value)
        }
    });
}
