// Copyright 2018 The go-klaytn Authors
// This file is part of the go-klaytn library.
//
// The go-klaytn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-klaytn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-klaytn library. If not, see <http://www.gnu.org/licenses/>.

package impl

import (
	"bytes"
	"fmt"
	"github.com/ground-x/go-gxplatform/cmd/utils"
	"os"
	"text/template"
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
import "github.com/ground-x/go-gxplatform/client"
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
	gxplatformFlag      = flag.String(gxpName, "ws://127.0.0.1:8546", "the klay client address")
	portFlag            = flag.String(portName, "127.0.0.1:5555", "server port")
	privateKeyFlag      = flag.String(privateKeyName, "", "deployer's private key")
	contractAddressFlag = flag.String(contractAddressName, "", "contract address")
)

func main() {
	flag.Parse()

	viper.BindPFlags(flag.CommandLine)
	viper.AutomaticEnv() // read in environment variables that match

	klay := viper.GetString(gxpName)
	if klay == "" {
		fmt.Printf("No klay client specified\n")
		os.Exit(-1)
	}

	port := viper.GetString(portName)
	if port == "" {
		fmt.Printf("No listen port specified\n")
		os.Exit(-1)
	}

	// connect to klay client
	conn, err := client.Dial(klay)
	if err != nil {
		fmt.Printf("Failed to connect klay: %v\n", err)
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
