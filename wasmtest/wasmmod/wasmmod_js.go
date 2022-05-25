// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The wasmmod is a Tailscale-in-wasm proof of concept.
//
// See ../index.html and ../term.js for how it ties together.
package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"html"
	"log"
	"net"
	"strings"
	"syscall/js"
	"time"

	"golang.org/x/crypto/ssh"
	"inet.af/netaddr"
	"tailscale.com/ipn"
	"tailscale.com/ipn/ipnserver"
	"tailscale.com/net/netns"
	"tailscale.com/net/tsdial"
	"tailscale.com/safesocket"
	"tailscale.com/types/logger"
	"tailscale.com/wgengine"
	"tailscale.com/wgengine/netstack"
)

func main() {
	netns.SetEnabled(false)
	var logf logger.Logf = log.Printf

	dialer := new(tsdial.Dialer)
	eng, err := wgengine.NewUserspaceEngine(logf, wgengine.Config{
		Dialer: dialer,
	})
	if err != nil {
		log.Fatal(err)
	}

	tunDev, magicConn, dnsManager, ok := eng.(wgengine.InternalsGetter).GetInternals()
	if !ok {
		log.Fatalf("%T is not a wgengine.InternalsGetter", eng)
	}
	ns, err := netstack.Create(logf, tunDev, eng, magicConn, dialer, dnsManager)
	if err != nil {
		log.Fatalf("netstack.Create: %v", err)
	}
	ns.ProcessLocalIPs = true
	ns.ProcessSubnets = true
	if err := ns.Start(); err != nil {
		log.Fatalf("failed to start netstack: %v", err)
	}
	dialer.UseNetstackForIP = func(ip netaddr.IP) bool {
		_, ok := eng.PeerForIP(ip)
		return ok
	}
	dialer.NetstackDialTCP = func(ctx context.Context, dst netaddr.IPPort) (net.Conn, error) {
		return ns.DialContextTCP(ctx, dst)
	}

	doc := js.Global().Get("document")
	state := doc.Call("getElementById", "state")
	netmapEle := doc.Call("getElementById", "netmap")
	loginEle := doc.Call("getElementById", "loginURL")

	var store ipn.StateStore = new(jsStateStore)
	srv, err := ipnserver.New(log.Printf, "some-logid", store, eng, dialer, nil, ipnserver.Options{
		SurviveDisconnects: true,
	})
	if err != nil {
		log.Fatalf("ipnserver.New: %v", err)
	}
	lb := srv.LocalBackend()

	state.Set("innerHTML", "ready")

	lb.SetNotifyCallback(func(n ipn.Notify) {
		log.Printf("NOTIFY: %+v", n)
		if n.State != nil {
			state.Set("innerHTML", fmt.Sprint(*n.State))
			switch *n.State {
			case ipn.Running, ipn.Starting:
				loginEle.Set("innerHTML", "")
			}
		}
		if nm := n.NetMap; nm != nil {
			var buf bytes.Buffer
			fmt.Fprintf(&buf, "<p>Name: <b>%s</b></p>\n", html.EscapeString(nm.Name))
			fmt.Fprintf(&buf, "<p>Addresses: ")
			for i, a := range nm.Addresses {
				if i == 0 {
					fmt.Fprintf(&buf, "<b>%s</b>", a.IP())
				} else {
					fmt.Fprintf(&buf, ", %s", a.IP())
				}
			}
			fmt.Fprintf(&buf, "</p>")
			fmt.Fprintf(&buf, "<p>Machine: <b>%v</b>, %v</p>\n", nm.MachineStatus, nm.MachineKey)
			fmt.Fprintf(&buf, "<p>Nodekey: %v</p>\n", nm.NodeKey)
			fmt.Fprintf(&buf, "<hr><table>")
			for _, p := range nm.Peers {
				var ip string
				if len(p.Addresses) > 0 {
					ip = p.Addresses[0].IP().String()
				}
				fmt.Fprintf(&buf, "<tr><td>%s</td><td>%s</td></tr>\n", ip, html.EscapeString(p.Name))
			}
			fmt.Fprintf(&buf, "</table>")
			netmapEle.Set("innerHTML", buf.String())
		}
		if n.BrowseToURL != nil {
			js.Global().Call("browseToURL", *n.BrowseToURL)
		}
	})

	start := func() {
		err := lb.Start(ipn.Options{
			StateKey: "wasm",
			UpdatePrefs: &ipn.Prefs{
				// go run ./cmd/trunkd/  -remote-url=https://controlplane.tailscale.com
				//ControlURL:       "http://tsdev:8080",
				ControlURL:       ipn.DefaultControlURL,
				RouteAll:         false,
				AllowSingleHosts: true,
				WantRunning:      true,
				Hostname:         "wasm",
			},
		})
		log.Printf("Start error: %v", err)

	}

	js.Global().Set("startClicked", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		go start()
		return nil
	}))

	js.Global().Set("logoutClicked", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Printf("Logout clicked")
		if lb.State() == ipn.NoState {
			log.Printf("Backend not running")
			return nil
		}
		go lb.Logout()
		return nil
	}))

	js.Global().Set("startLoginInteractive", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		log.Printf("State: %v", lb.State)

		go func() {
			if lb.State() == ipn.NoState {
				start()
			}
			lb.StartLoginInteractive()
		}()
		return nil
	}))

	js.Global().Set("runSSH", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) < 2 {
			log.Printf("missing args")
			return nil
		}
		go func() {
			onDone := args[1]
			defer onDone.Invoke() // re-print the prompt

			line := args[0].String()
			f := strings.Fields(line)
			host := f[1]

			term := js.Global().Get("theTerminal")
			writeError := func(label string, err error) {
				term.Call("write", fmt.Sprintf("%s Error: %v\r\n", label, err))
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			c, err := dialer.UserDial(ctx, "tcp", net.JoinHostPort(host, "22"))
			if err != nil {
				writeError("Dial", err)
				return
			}
			defer c.Close()

			config := &ssh.ClientConfig{
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}

			sshConn, _, _, err := ssh.NewClientConn(c, host, config)
			if err != nil {
				writeError("SSH Connection", err)
				return
			}
			defer sshConn.Close()
			term.Call("write", "SSH Connected\r\n")

			sshClient := ssh.NewClient(sshConn, nil, nil)
			defer sshClient.Close()

			session, err := sshClient.NewSession()
			if err != nil {
				writeError("SSH Session", err)
				return
			}
			term.Call("write", "Session Established\r\n")
			defer session.Close()

			stdin, err := session.StdinPipe()
			if err != nil {
				writeError("SSH Stdin", err)
				return
			}

			session.Stdout = termWriter{term}
			session.Stderr = termWriter{term}

			term.Set("_onDataHook", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				input := args[0].String()
				_, err := stdin.Write([]byte(input))
				if err != nil {
					writeError("Write Input", err)
				}
				return nil
			}))
			defer func() {
				term.Delete("_onDataHook")
			}()

			err = session.RequestPty("xterm", term.Get("rows").Int(), term.Get("cols").Int(), ssh.TerminalModes{})

			if err != nil {
				writeError("Pseudo Terminal", err)
				return
			}

			err = session.Shell()
			if err != nil {
				writeError("Shell", err)
				return
			}

			err = session.Wait()
			if err != nil {
				writeError("Exit", err)
				return
			}
		}()
		return nil
	}))

	ln, _, err := safesocket.Listen("", 0)
	if err != nil {
		log.Fatal(err)
	}

	err = srv.Run(context.Background(), ln)
	log.Fatalf("ipnserver.Run exited: %v", err)
}

type termWriter struct {
	o js.Value
}

func (w termWriter) Write(p []byte) (n int, err error) {
	r := bytes.Replace(p, []byte("\n"), []byte("\n\r"), -1)
	w.o.Call("write", string(r))
	return len(p), nil
}

type jsStateStore struct{}

func (_ *jsStateStore) ReadState(id ipn.StateKey) ([]byte, error) {
	jsValue := js.Global().Call("getState", string(id))
	if jsValue.String() == "" {
		return nil, ipn.ErrStateNotExist
	}
	return hex.DecodeString(jsValue.String())
}

func (_ *jsStateStore) WriteState(id ipn.StateKey, bs []byte) error {
	js.Global().Call("setState", string(id), hex.EncodeToString(bs))
	return nil
}
