require('../semantic/dist/components/api')
require('../semantic/dist/components/checkbox')
require('../semantic/dist/components/dimmer')
require('../semantic/dist/components/dropdown')
require('../semantic/dist/components/modal')
require('../semantic/dist/components/search')
require('../semantic/dist/components/toast')
require('../semantic/dist/components/transition')
import '../semantic/dist/semantic.min.css'
import "xterm/css/xterm.css";
import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';
import { FitAddon } from 'xterm-addon-fit';
import * as XtermWebfont from 'xterm-webfont'
import * as $ from 'jquery'

let terminal: Terminal

$(document).ready(function () {
})


$('#btn').on('click', null, function (e) {
    let appName = "TestApp"

    let modal = `
    <div class="ui modal" id="start-app-modal">
        <i class="close icon"></i>
        <div class="header">
            Log for application ${appName}
        </div>
        <div id="terminal"></div>
    </div>`
    $(modal).modal({
        allowMultiple: false,
        onVisible: function (e) {
            console.log("run open");
            terminal = new Terminal({
                cursorBlink: false,
                tabStopWidth: 4,
                lineHeight: 1,
                fontSize: 11
            });
            const socket = new WebSocket('ws://localhost:9090/echo')
            const attachAddon = new AttachAddon(socket)
            const fitAddon = new FitAddon()
        
            // Attach the socket to term
            terminal.loadAddon(attachAddon)
            terminal.loadAddon(fitAddon)
            
            terminal.open(document.getElementById('terminal') as HTMLElement)
            fitAddon.fit()

            terminal.write('Hello from \x1B[1;3;31mxterm.js\x1B[0m $ ')
        },
        onHidden: function (e) {
            console.log("run dispose");
            terminal.dispose()
            $(this).off(e);    // remove this listener
            $(this).modal('destroy');  // take down the modal object
            $(this).remove();    // remove the modal element, at last.
        }
    }).modal('show')
})