import "xterm/css/xterm.css";
import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';

const terminal = new Terminal({
    cursorBlink: false,
    tabStopWidth: 4,
    disableStdin: true,
    fontSize: 13,
    lineHeight: 1,
    theme: {
        background: "#151515",
    },

});
const socket = new WebSocket('ws://localhost:9090/echo');
const attachAddon = new AttachAddon(socket);

// Attach the socket to term
terminal.loadAddon(attachAddon);

terminal.open(document.getElementById('terminal'));

// socket.send("HGFGFGHF")
