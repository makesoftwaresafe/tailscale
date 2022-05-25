// Hacked up version of https://xtermjs.org/js/demo.js
// for now.

import { Terminal } from 'xterm';
import { WebglAddon } from 'xterm-addon-webgl';
import QRCode from "qrcode";

function main() {
  // Custom theme to match style of xterm.js logo
  var baseTheme = {
    foreground: '#F8F8F8',
    background: '#2D2E2C',
    selection: '#5DA5D533',
    black: '#1E1E1D',
    brightBlack: '#262625',
    red: '#CE5C5C',
    brightRed: '#FF7272',
    green: '#5BCC5B',
    brightGreen: '#72FF72',
    yellow: '#CCCC5B',
    brightYellow: '#FFFF72',
    blue: '#5D5DD3',
    brightBlue: '#7279FF',
    magenta: '#BC5ED1',
    brightMagenta: '#E572FF',
    cyan: '#5DA5D5',
    brightCyan: '#72F0FF',
    white: '#F8F8F8',
    brightWhite: '#FFFFFF'
  };

  const termContainerNode = document.querySelector('.term .inner');
  var term = new Terminal({
    fontFamily: '"Cascadia Code", Menlo, monospace',
    theme: baseTheme,
    cursorBlink: true
  });
  term.open(termContainerNode);
  globalThis.theTerminal = term;

  var isWebglEnabled = false;
  try {
    const webgl = new WebglAddon();
    term.loadAddon(webgl);
    isWebglEnabled = true;
  } catch (e) {
    console.warn('WebGL addon threw an exception during load', e);
  }

  // Cancel wheel events from scrolling the page if the terminal has scrollback
  termContainerNode.addEventListener('wheel', e => {
    if (term.buffer.active.baseY > 0) {
      e.preventDefault();
    }
  });

  function runFakeTerminal() {
    if (term._initialized) {
      return;
    }

    term._initialized = true;

    term.prompt = () => {
      term.write('\r\n$ ');
    };

    // TODO: Use a nicer default font
    term.writeln('Tailscale js/wasm demo; try running `help`.');
    prompt(term);

    term.onData(e => {
      if (term._onDataHook) {
        term._onDataHook(e);
        return;
      }
      switch (e) {
        case '\u0003': // Ctrl+C
          term.write('^C');
          prompt(term);
          break;
        case '\r': // Enter
          runCommand(term, command);
          command = '';
          break;
        case '\u007F': // Backspace (DEL)
          // Do not delete the prompt
          if (term._core.buffer.x > 2) {
            term.write('\b \b');
            if (command.length > 0) {
              command = command.substr(0, command.length - 1);
            }
          }
          break;
        default: // Print all other characters for demo
          if (e >= String.fromCharCode(0x20) && e <= String.fromCharCode(0x7B) || e >= '\u00a0') {
            command += e;
            term.write(e);
          }
      }
    });
  }

  function prompt(term) {
    command = '';
    term.write('\r\n$ ');
  }

  var command = '';
  var commands = {
    help: {
      f: () => {
        term.writeln([
          'Welcome to Tailscale js/wasm! Try some of the commands below.',
          '',
          ...Object.keys(commands).map(e => `  ${e.padEnd(10)} ${commands[e].description}`)
        ].join('\n\r'));
        prompt(term);
      },
      description: 'Prints this help message',
    },
    ssh: {
      f: (line) => {
        runSSH(line, function () { term.prompt(term) });
      },
      description: 'SSH to a Tailscale peer'
    },
  };

  function runCommand(term, text) {
    const command = text.trim().split(' ')[0];
    if (command.length > 0) {
      term.writeln('');
      if (command in commands) {
        commands[command].f(text);
        return;
      }
      term.writeln(`${command}: command not found`);
    }
    prompt(term);
  }

  runFakeTerminal();
}

globalThis.updateNetMap = function(netMapStr) {
  const netMap = JSON.parse(netMapStr);
  const netmapMode = document.getElementById("netMap");
  netmapMode.textContent = JSON.stringify(netMap, null, 2);
}

globalThis.browseToURL = async function(url) {
  const loginNode = document.getElementById("loginURL");
  loginNode.innerHTML = "";
  const linkNode = document.createElement("a");
  linkNode.href = url;
  linkNode.target = "_blank";
  linkNode.textContent = url;
  loginNode.appendChild(linkNode);

  try {
    const dataURL = await QRCode.toDataURL(url, {width: 512});
    linkNode.appendChild(document.createElement("br"));
    const imageNode = document.createElement("img");
    imageNode.src = dataURL;
    imageNode.width = 256;
    imageNode.height = 256;
    imageNode.border = "0";
    linkNode.appendChild(imageNode);
  } catch (err) {
    console.error("Could not generate QR code:", err);
  }
}

// Used by jsStateStore to persist IPN state
window.setState = (id, value) => window.sessionStorage[`ipn-state-${id}`] = value;
window.getState = (id) => window.sessionStorage[`ipn-state-${id}`] || "";

main();
