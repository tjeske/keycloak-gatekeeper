require('../semantic/dist/components/api')
require('../semantic/dist/components/checkbox')
require('../semantic/dist/components/dimmer')
require('../semantic/dist/components/dropdown')
require('../semantic/dist/components/modal')
require('../semantic/dist/components/search')
require('../semantic/dist/components/form')
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

/**
 * Initialize drop down list for app templates.
 */
export function initTemplateDropDown() {
    // get templates
    $.ajax({
        url: "/udesk/getTemplates",
        success: function (result: { data: { name: string; uuid: string}[] }) {
            let templates = result.data
            let values = templates.map(it => {
                let values: { [key: string]: string; } = {}
                values['name'] = it.name
                values['value'] = it.name
                return values
            });
            ($('#app-dropdown') as any).dropdown('change values', values);
            if (templates.length > 0) {
                $('#app-dropdown').dropdown('set selected', templates[0].name);
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

