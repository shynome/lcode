/*
Copyright © 2024 shynome <shynome@gmail.com>
*/
package cmd

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/shynome/err0/try"
	"github.com/spf13/cobra"
	"golang.org/x/net/webdav"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var args struct {
	connect string
	webdav  bool
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lcode",
	Short: "lcode",
	Long:  `lcode`,
	Run: func(cmd *cobra.Command, files []string) {
		ctx := context.Background()
		hostname := try.To1(os.Hostname())
		macHash := getMacHashTry()
		socket, resp := try.To2(websocket.Dial(ctx, args.connect, &websocket.DialOptions{
			Subprotocols: []string{"webdav", hostname, macHash},
		}))

		if len(files) == 0 {
			files = []string{"."}
		}
		wd := try.To1(os.Getwd())
		var allow = make([]string, len(files))
		for i, f := range files {
			if strings.HasPrefix(f, "/") {
				allow[i] = filepath.Join("/", f)
			} else {
				allow[i] = filepath.Join(wd, f)
			}
		}

		func() {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			try.To(wsjson.Write(ctx, socket, allow))
		}()

		conn := websocket.NetConn(ctx, socket, websocket.MessageBinary)
		webdavHost := resp.Header.Get("X-Webdav-Host")
		sess := try.To1(yamux.Client(conn, nil))
		defer sess.Close()

		rpcConn := try.To1(sess.Open())
		client := rpc.NewClient(rpcConn)

		rls := NewRemoteLockSystem(client)

		var h http.Handler = &webdav.Handler{
			FileSystem: webdav.Dir("/"),
			LockSystem: rls,
		}
		for _, f := range allow {
			if true {
				link := fmt.Sprintf("vscode://lcode.hub/%s%s", webdavHost, f)
				if stat, err := os.Stat(f); err == nil && !stat.IsDir() {
					link += "#file"
				}
				log.Println(link)
			}
			if args.webdav {
				link := fmt.Sprintf("webdav://%s%s", webdavHost, f)
				log.Println(link)
			}
		}
		http.Serve(sess, h)
	},
}

func getMacHashTry() string {
	ifaces, _ := net.Interfaces()
	if len(ifaces) == 0 {
		home := try.To1(os.UserHomeDir())
		lcodeDir := filepath.Join(home, ".lcode")
		try.To(os.MkdirAll(lcodeDir, os.ModePerm))
		fakeMacFile := filepath.Join(lcodeDir, "fake-mac")
		fakeMac, _ := os.ReadFile(fakeMacFile)
		if len(fakeMac) == 0 {
			fakeMac = make([]byte, 6)
			try.To1(rand.Read(fakeMac))
			try.To(os.WriteFile(fakeMacFile, fakeMac, os.ModePerm))
		}
		ifaces = []net.Interface{
			{Flags: net.FlagUp, HardwareAddr: fakeMac},
		}
	}
	hasher := fnv.New32a()
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp != 0 && iface.HardwareAddr != nil {
			try.To1(hasher.Write(iface.HardwareAddr))
			break
		}
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.v3.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&args.connect, "connect", "c", "ws://127.0.0.1:4349", "")
	rootCmd.Flags().BoolVarP(&args.webdav, "webdav", "w", false, "")
}
