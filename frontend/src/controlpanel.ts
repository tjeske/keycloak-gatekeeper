import { addParameter, initAppDropDown, initTemplateDropDown, showRemoveContainerModal, App } from 'shared'
import "xterm/css/xterm.css";
import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';
import { FitAddon } from 'xterm-addon-fit';
import * as XtermWebfont from 'xterm-webfont'
require('typeface-courier-prime')

function showStartNewAppModal(templateName: string, params: Map<string, string>, table: DataTables.Api) {

    let rows = []
    for (let [key, value] of Object.entries(params)) {
        let row = `
        <tr>
            <td>
                <div class="ui"><strong>${key}</strong></div>
            </td>
            <td>
                <div class="ui black input"><input type="text" name="${key}" value="${value}"></div>
            </td>
        </tr>`
        rows.push(row)
    }

    let modal = `
    <div class="ui tiny modal" id="start-new-app-modal">
        <div class="content">
            <p>Start new application with template <strong>${templateName}</strong>?</p>
            <form class="tagForm" id="tag-form" action="/startApp" method="post" enctype="multipart/form-data">
                <input type="hidden" name="templateName" value="${templateName}">
                <table class="ui very basic collapsing table" id="params-table">
                    <tbody id="params-table-body">
                        <tr>
                            <td>
                                <div class="ui"><strong>name</strong></div>
                            </td>
                            <td>
                                <div class="ui black input"><input type="text" name="name" value=""></div>
                            </td>
                        </tr>
                        ${rows.join("")}
                    </tbody>
                </table>
            </form>
        </div>
        <div class="actions">
            <div class="ui negative button">
                Cancel
            </div>
            <div class="ui positive right labeled icon button">
                Start
                <i class="play icon"></i>
            </div>
        </div>
    </div>`
    $(modal).modal({
        onApprove: function (e) {
            $.ajax({
                url: "/udesk/startApp",
                type: "GET",
                data: $('form.tagForm').serialize(),
                success: function (result) {
                    ($('body') as any).toast({
                        class: 'success',
                        message: `Application started successfully!`
                    });
                    table.ajax.reload();
                },
                error: function (result: { responseText: string }) {
                    let msg = ""
                    if (result != null) {
                        msg = `Cannot create new application (error: ${result.responseText})!`
                    } else {
                        msg = `Cannot create new application!`
                    }
                    
                    ($('body') as any).toast({
                        class: 'error',
                        message: msg
                    });
                }
            });
        },
    }).modal('show')
}

function initStartNewAppButton(table: DataTables.Api) {
    // button remove
    $('#start-new-app-btn').on('click', null, function (e) {

        let templateName = $('#app-dropdown').dropdown('get value');
        $.ajax({
            url: "/udesk/getTemplates",
            success: function (result: { data: App[] }) {
                let config = result.data
                let element = config.find(it => 'name' in it && it.name == templateName)

                if (element != undefined && 'params' in element  && element.params != null ) {
                    showStartNewAppModal(templateName, element.params, table)
                } else {
                    showStartNewAppModal(templateName, new Map, table)
                }
            },
            error: function (result) {
                console.log(result);
                
                ($('body') as any).toast({
                    class: 'error',
                    message: `Cannot get template list from server`
                });
            }
        });
    });
}

/**
 * Initialize the app status panel.
 */
function initAppStatusView(): DataTables.Api {
    // init table
    let table = $('#app-status-table').DataTable({
        "ajax": {
            "url": '/udesk/dockerStatus'
        },
        "columnDefs": [
            {
                // name
                "render": function (name: string, type: string, row: string[]) {
                    let uuid = row[4]
                    return `<a href="/udesk/switchApp/${uuid}">${name}</a><span style="float:right;"><i class="red icon delete app-remove-btn" data-container-name="${name}" data-container-id="${uuid}"></i></span>`
                },
                "targets": 0
            },
            {
                // controls
                "render": function (state: string, type: string, row: string[]) {
                    let name = row[0]
                    let containerId = row[4]
                    switch (state) {
                        case "running":
                            return `${state} <i class="orange icon pause app-pause-btn" data-container-name="${name}" data-container-id="${containerId}"></i><div id="container"></div>`
                        case "paused":
                            return `${state} <i class="green icon play app-unpause-btn" data-container-name="${name}" data-container-id="${containerId}"></i>`
                        default:
                            return state
                    }
                },
                "targets": 3
            },
            { "visible": false, "targets": [4] }
        ],
        "deferRender": true,
        "info": false,
        "paging": false,
        "searching": false
    });
    // button unpause
    $('#app-status-table').on('click', '.app-unpause-btn', function (e) {
        let target = $(e.target)
        let containerId = target.data("container-id")
        let containerName = target.data("container-name")
        $.ajax({
            url: "/udesk/unpauseApp/" + containerId,
            success: function (result) {
                ($('body') as any).toast({
                    class: 'success',
                    message: `Successfully started container <strong>${containerName}</strong>`
                });
                table.ajax.reload();
            },
            error: function (result) {
                ($('body') as any).toast({
                    class: 'error',
                    message: `Starting container <strong>${containerName}</strong> failed
                    <strong>${result}</strong>
                    `
                });
            }
        });
    });
    // button remove
    $('#app-status-table').on('click', '.app-remove-btn', function (e) {
        let target = $(e.target)
        showRemoveContainerModal(target.data("container-name"), target.data("container-id"), table)
    });
    // button pause
    $('#app-status-table').on('click', '.app-pause-btn', function (e) {
        let target = $(e.target)
        let containerId = target.data("container-id")
        let containerName = target.data("container-name")
        $.ajax({
            url: "/udesk/pauseApp/" + containerId,
            success: function (result) {
                ($('body') as any).toast({
                    class: 'success',
                    message: `Successfully paused container <strong>${containerName}</strong>!`
                });
                table.ajax.reload();
            },
            error: function (result) {
                ($('body') as any).toast({
                    class: 'error',
                    message: `Pausing container <strong>${containerName}</strong> failed!`
                });
            }
        });
    });
    // update
    setInterval(function () {
        table.ajax.reload();
    }, 1000);

    return table
}

$(document).ready(function () {
    initTemplateDropDown()
    let table = initAppStatusView()

    initStartNewAppButton(table)

    const terminal = new Terminal({
        cursorBlink: false,
        tabStopWidth: 4,
        // disableStdin: true,
        lineHeight: 1,
        // windowsMode: ['Windows', 'Win16', 'Win32', 'WinCE'].indexOf(navigator.platform) >= 0,
    // convertEol: true,
    //fontFamily: `'Lato', 'Courier Prime', 'Lato'`,
    fontFamily: `'Courier Prime'`,
    fontSize: 11,
    // fontWeight: 400,
    // rendererType: "canvas" // canvas 或者 dom
    
    });
    const socket = new WebSocket('ws://localhost:3000/udesk/echo')
    const attachAddon = new AttachAddon(socket)
    const fitAddon = new FitAddon()

    // Attach the socket to term
    terminal.loadAddon(attachAddon)
    terminal.loadAddon(fitAddon)
    terminal.loadAddon(new XtermWebfont())
    
    terminal.loadWebfontAndOpen(document.getElementById('terminal') as HTMLElement)
    fitAddon.fit()
})