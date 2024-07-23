/*
Copyright © 2024 shynome <shynome@gmail.com>
*/
package cmd

import (
	"context"
	"encoding/hex"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path/filepath"
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
			allow[i] = filepath.Join(wd, f)
		}

		func() {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			try.To(wsjson.Write(ctx, socket, allow))
		}()

		conn := websocket.NetConn(ctx, socket, websocket.MessageBinary)
		webdavHost := resp.Header.Get("X-Webdav-Host")
		log.Println(webdavHost)
		sess := try.To1(yamux.Client(conn, nil))
		defer sess.Close()

		rpcConn := try.To1(sess.Open())
		client := rpc.NewClient(rpcConn)

		rls := NewRemoteLockSystem(client)

		var h http.Handler = &webdav.Handler{
			FileSystem: webdav.Dir("/"),
			LockSystem: rls,
		}
		http.Serve(sess, h)
	},
}

func getMacHashTry() string {
	ifaces := try.To1(net.Interfaces())
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
