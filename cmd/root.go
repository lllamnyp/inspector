/*
Copyright © 2024 Timofei Larkin <lllamnyp@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	stdUrl "net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/lllamnyp/inspector/pkg/handler"
	"github.com/lllamnyp/inspector/pkg/url"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "inspector",
	Short: "MITM proxy to inspect requests and responses",
	Long: `Inspector can intercept HTTP(s) requests from your apps
between one annother and to remote servers, log them and
the respective responses and help you debug that pesky IdP
integration or some other misbehaving server.

Usage:

  Inspector consumes a map[string]string under the --proxy flag:

    ./inspector --proxy=http://some-alias.local/=https://example.com/ \
      --proxy=https://other.example.com/=https://other.example.com/ \
      --proxy=http://my-port.local:8080/=http://custom-port.com:9090/

  Use --cert and --key flags, when listening with TLS:

    ./inspector --cert cert.pem --key key.pem \ ...

  Use --cacert and --cakey to autogenerate the right certs for the job (WIP):

    ./inspector --cacert ca.pem --cakey private.pem \ ...`,
	Run: func(cmd *cobra.Command, args []string) {
		proxy, err := cmd.Flags().GetStringToString("proxy")
		if err != nil {
			panic(err)
		}
		portToScheme := make(map[string]string)
		portToHostPathMap := make(map[string]map[url.URL]url.URL)
		for entry, backend := range proxy {
			e, err := url.Parse(entry)
			if err != nil {
				fmt.Printf("cannot parse url %s, err: %s", entry, err)
				continue
			}
			b, err := url.Parse(backend)
			if err != nil {
				fmt.Printf("cannot parse url %s, err: %s", backend, err)
				continue
			}
			if _, ok := portToScheme[e.Port]; !ok {
				portToScheme[e.Port] = e.Scheme
			}
			if portToScheme[e.Port] != e.Scheme {
				fmt.Println("http and https servers listening on a single port are not allowed")
				return
			}
			if _, ok := portToHostPathMap[e.Port]; !ok {
				portToHostPathMap[e.Port] = make(map[url.URL]url.URL)
			}
			// TODO: validate uniqueness of hostpath patterns
			portToHostPathMap[e.Port][e] = b
		}
		srvs := make(map[string]*http.Server)
		for port := range portToScheme {
			srvs[port] = &http.Server{Handler: http.NewServeMux()}
			for e, b := range portToHostPathMap[port] {
				backendUrl, _ := stdUrl.Parse(b.Scheme + "://" + b.Hostname + ":" + b.Port + b.Path)
				proxyHandler := &httputil.ReverseProxy{Rewrite: func(pr *httputil.ProxyRequest) { pr.SetURL(backendUrl) }}
				hdl := handler.LogAndRewrite(e)(proxyHandler.ServeHTTP)
				srvs[port].Addr = ":" + port
				srvs[port].Handler.(*http.ServeMux).HandleFunc(e.Hostname+e.Path, hdl)
			}
			go srvs[port].ListenAndServe()
		}
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		wg := &sync.WaitGroup{}
		for _, srv := range srvs {
			wg.Add(1)
			go func(srv *http.Server) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				defer wg.Done()
				srv.Shutdown(ctx)
			}(srv)
		}
		wg.Wait()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.inspector.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().StringToString("proxy", map[string]string{}, "A map of URLs to listen on (keys) and the URLs to proxy to (values).")
}
