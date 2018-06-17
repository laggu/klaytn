package impl

import (
	"fmt"
	"os"
	"bytes"
	"text/template"
	"github.com/ground-x/go-gxplatform/cmd/utils"
)

type Server struct {
	ContractName   string
	ProjectPackage string
}

var ServerTemplate string = `package main

import "context"
import "fmt"
import "log"
import "math/big"
import "net"
import "os"

import "github.com/ground-x/go-gxplatform/accounts/abi/bind"
import "github.com/ground-x/go-gxplatform/common"
import "github.com/ground-x/go-gxplatform/crypto"
import "github.com/ground-x/go-gxplatform/gxpclient"
import {{ .ContractName }} "{{ .ProjectPackage }}"
import flag "github.com/spf13/pflag"
import "github.com/spf13/viper"
import "google.golang.org/grpc"

const (
	gxpName             = "gxplatform"
	portName            = "port"
	privateKeyName      = "private_key"
	contractAddressName = "contract_address"
)

var (
	gxplatformFlag      = flag.String(gxpName, "ws://127.0.0.1:8546", "the gxp client address")
	portFlag            = flag.String(portName, "127.0.0.1:5555", "server port")
	privateKeyFlag      = flag.String(privateKeyName, "", "deployer's private key")
	contractAddressFlag = flag.String(contractAddressName, "", "contract address")
)

func main() {
	flag.Parse()

	viper.BindPFlags(flag.CommandLine)
	viper.AutomaticEnv() // read in environment variables that match

	gxp := viper.GetString(gxpName)
	if gxp == "" {
		fmt.Printf("No gxp client specified\n")
		os.Exit(-1)
	}

	port := viper.GetString(portName)
	if port == "" {
		fmt.Printf("No listen port specified\n")
		os.Exit(-1)
	}

	// connect to gxp client
	conn, err := gxpclient.Dial(gxp)
	if err != nil {
		fmt.Printf("Failed to connect gxp: %v\n", err)
		os.Exit(-1)
	}

	privateKey := viper.GetString(privateKeyName)

	// Deploy contracts
	var addr common.Address
	if privateKey != "" {
		// set up auth
		key, err := crypto.ToECDSA(common.Hex2Bytes(privateKey))
		if err != nil {
			fmt.Printf("Failed to get private key: %v\n", err)
			os.Exit(-1)
		}
		auth := bind.NewKeyedTransactor(key)
		auth.GasLimit = uint64(9000000000000)
		// get nonce
		nonce, err := conn.NonceAt(context.Background(), auth.From, nil)
		if err != nil {
			fmt.Printf("Failed to get nonce: %v\n", err)
			os.Exit(-1)
		}
		auth.Nonce = big.NewInt(int64(nonce))

		addr, _, _, err = {{ .ContractName }}.Deploy{{ .ContractName }}(auth, conn)
		if err != nil {
			fmt.Printf("Failed to deploy contract: %v\n", err)
			os.Exit(-1)
		}

		fmt.Printf("Deployed contract: %v\n", addr.Hex())
	} else {
		address := viper.GetString(contractAddressName)
		if address == "" {
			fmt.Printf("No contract address specified\n")
			os.Exit(-1)
		}
		addr = common.HexToAddress(address)
	}

	s := grpc.NewServer()
	{{ .ContractName }}.Register{{ .ContractName }}Server(s, {{ .ContractName }}.New{{ .ContractName }}Server(addr, conn, nil))

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s.Serve(lis)
}
`

func (s *Server) Write(filepath, filename string) {
	implTemplate, err := template.New("server").Parse(ServerTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		os.Exit(-1)
	}
	result := new(bytes.Buffer)
	err = implTemplate.Execute(result, s)
	if err != nil {
		fmt.Printf("Failed to render template: %v\n", err)
		os.Exit(-1)
	}
	utils.WriteFile(result.String(), filepath, filename)
}

