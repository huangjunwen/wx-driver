// Copyright © 2018 Huang junwen
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/huangjunwen/wxdriver/utils"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

var (
	cfgFile string
	verbose bool
	timeout string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wxdrv",
	Short: "Wechat API utilities",
	Long:  ``,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return initHTTPClient()
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
		//log.Fatal(err)
	}
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.wxdrv.yaml)")

	// 一些全局选项
	rootCmd.PersistentFlags().StringP("appid", "a", "", "Wechat app id [required|configurable]")
	viper.BindPFlag("appid", rootCmd.PersistentFlags().Lookup("appid"))

	rootCmd.PersistentFlags().StringVar(&timeout, "timeout", "5s", "Timeout duration")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "Verbose log")

	rootCmd.PersistentFlags().String("tls_cert", "", "TLS cert file or PEM block [configurable]")
	viper.BindPFlag("tls_cert", rootCmd.PersistentFlags().Lookup("tls_cert"))

	rootCmd.PersistentFlags().String("tls_key", "", "TLS key file or PEM block [configurable]")
	viper.BindPFlag("tls_key", rootCmd.PersistentFlags().Lookup("tls_key"))

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".wxdrv" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".wxdrv")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.ReadInConfig() // If a config file is found, read it in.

}

func initHTTPClient() error {

	var err error

	timeoutDur, err := time.ParseDuration(timeout)
	if err != nil {
		return err
	}

	var tlsConfig *tls.Config
	tlsKey := viper.GetString("tls_key")
	tlsCert := viper.GetString("tls_cert")

	if tlsKey != "" || tlsCert != "" {
		if tlsKey == "" || tlsCert == "" {
			return errors.New(`Both "tls_key" and "tls_cert" should be presented`)
		}

		tlsKeyPEM := []byte{}
		tlsCertPEM := []byte{}

		if strings.HasPrefix(tlsKey, "-----BEGIN") {
			tlsKeyPEM = []byte(tlsKey)
		} else {
			tlsKeyPEM, err = ioutil.ReadFile(tlsKey)
			if err != nil {
				return err
			}
		}

		if strings.HasPrefix(tlsCert, "-----BEGIN") {
			tlsCertPEM = []byte(tlsCert)
		} else {
			tlsCertPEM, err = ioutil.ReadFile(tlsCert)
			if err != nil {
				return err
			}
		}

		tlsConfig, err = utils.TLSConfig(tlsCertPEM, tlsKeyPEM, nil)
		if err != nil {
			return err
		}

	}

	var client utils.HTTPClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: timeoutDur,
	}

	if verbose {
		client = VerboseLogHTTPClient{client}
	}

	utils.DefaultHTTPClient = client
	return nil

}

type VerboseLogHTTPClient struct {
	client utils.HTTPClient
}

func (client VerboseLogHTTPClient) Do(req *http.Request) (*http.Response, error) {
	reqBody, err := utils.ReadAndReplaceRequestBody(req)
	if err != nil {
		return nil, err
	}
	log.Printf("HTTP Request: %s %+q %+q", req.Method, req.URL.String(), reqBody)

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}

	respBody, err := utils.ReadAndReplaceResponseBody(resp)
	if err != nil {
		return nil, err
	}
	log.Printf("HTTP Response: %s %s %+q", resp.Status, resp.Proto, respBody)

	return resp, nil
}

func dumpJSON(v interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(v)
}
