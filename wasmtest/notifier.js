import QRCode from "qrcode";
import {ssh} from "./ssh";

/**
 * @fileoverview Notification callback functions (bridged from ipn.Notify)
 */

const State = {
    NoState: 0,
    InUseOtherUser: 1,
    NeedsLogin: 2,
    NeedsMachineAuth: 3,
    Stopped: 4,
    Starting: 5,
    Running: 6,
};

 globalThis.notifyState = function(state) {
    const stateNode = document.getElementById("state");
    stateNode.textContent = `Unknown (${state})`;
    for (const [name, value] of Object.entries(State)) {
        if (state === value) {
            stateNode.textContent = name;
            break;
        }
    }
    if (state === State.Running || state === State.Starting) {
        const loginNode = document.getElementById("loginURL");
        loginNode.innerHTML = "";
    }
 };

 globalThis.notifyNetMap = function(netMapStr) {
    const netMap = JSON.parse(netMapStr);
    console.log("Received net map: " + JSON.stringify(netMap, null, 2));

    const peersNode = document.getElementById("peers");
    peersNode.innerHTML = "";

    for (const peer of netMap.peers) {
        if (!peer.hasSSHHostKeys) {
            continue;
        }
        const peerNode = document.createElement("div");
        peerNode.className = "peer";
        const nameNode = document.createElement("div");
        nameNode.className = "name";
        nameNode.textContent = peer.name;
        peerNode.appendChild(nameNode);

        const sshButtonNode = document.createElement("button");
        sshButtonNode.className = "ssh";
        sshButtonNode.addEventListener("click", function() {
            ssh(peer.name)
        });
        sshButtonNode.textContent = "SSH";
        peerNode.appendChild(sshButtonNode);

        peersNode.appendChild(peerNode);
    }
  };

  globalThis.notifyBrowseToURL = async function(url) {
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
  };
