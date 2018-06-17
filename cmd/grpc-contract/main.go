package main

import (
	"io/ioutil"
	"fmt"
	"os"
	"strings"
	"path"
	"github.com/ground-x/go-gxplatform/cmd/grpc-contract/internal/impl"
	"github.com/getamis/sirius/util"

	flag "github.com/spf13/pflag"
	parser "github.com/zpatrick/go-parser"
)

var (
	filepath string
	pbPath   string
	goTypes  []string
)

func init() {
	flag.StringArrayVar(&goTypes, "types", []string{}, "the go-binding files")
	flag.StringVar(&filepath, "path", ".", "path")
	flag.StringVar(&pbPath, "pb-path", ".", "pb path")
}

func main() {
	flag.Parse()

	// find all proto generated files
	pbInfos, err := ioutil.ReadDir(pbPath)
	if err != nil {
		fmt.Printf("Failed to list files: %v\n", err)
		os.Exit(-1)
	}
	var pbFiles []string
	for _, f := range pbInfos {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pb.go") {
			pbFiles = append(pbFiles, path.Join(pbPath, f.Name()))
		}
	}
	if len(pbFiles) == 0 {
		fmt.Printf("Cannot find the generated proto files\n")
		os.Exit(-1)
	}
	pbGoFiles, err := parser.ParseFiles(pbFiles)
	if err != nil {
		fmt.Printf("Failed to parse file: %v\n", err)
		os.Exit(-1)
	}
	if len(pbGoFiles) == 0 {
		fmt.Printf("Cannot get the go files\n")
		os.Exit(-1)
	}

	// create a map to search quickly
	pbFilesMap := make(map[string]*parser.GoFile)
	for i, p := range pbFiles {
		pbFilesMap[p] = pbGoFiles[i]
	}

	// find pb package
	var pbPackage string
	if filepath == pbPath {
		pbPackage = ""
	} else {
		pbPackage = pbGoFiles[0].Package
	}

	// find contract package
	pack := path.Base(filepath)
	for _, goType := range goTypes {
		file := path.Join(filepath, goType+".go")
		goBindingFile, err := parser.ParseSingleFile(file)
		if err != nil {
			fmt.Printf("Failed to parse file: %v\n", err)
			os.Exit(-1)
		}

		contract := impl.NewContract(pack, pbPackage,
			util.ToCamelCase(goType),
			[]string{
				file,
				path.Join(pbPath, goType+".pb.go"),
				path.Join(pbPath, "messages.pb.go"),
			})

		// find the corresponding server interface
		f, ok := pbFilesMap[contract.Sources[1]]
		if !ok {
			fmt.Printf("Failed to load corresponding source file(%s) for service %v\n", contract.Sources[1], goType)
			os.Exit(-1)
		}
		var serverInterface *parser.GoInterface
		for _, i := range f.Interfaces {
			if contract.IsServerInterface(i.Name) {
				serverInterface = i
				break
			}
		}
		if serverInterface == nil {
			fmt.Printf("Failed to load corresponding server interface for service %v\n", goType)
			os.Exit(-1)
		}

		// find the corresponding server interface
		f, ok = pbFilesMap[contract.Sources[2]]
		if !ok {
			fmt.Printf("Failed to find corresponding server interface(%s) for service %v\n", contract.Sources[2], goType)
			os.Exit(-1)
		}

		// Try to find the grpc server intreface
		for _, m := range serverInterface.Methods {
			// Find request struct
			requestStructName := m.Params[1].Type[1:]
			request := findGoStruct(requestStructName, f)
			if request == nil {
				fmt.Printf("Failed to load corresponding request struct in method %v\n", m.Name)
				os.Exit(-1)
			}

			// Find response struct
			responseStructName := m.Results[0].Type[1:]
			response := findGoStruct(responseStructName, f)
			if response == nil {
				fmt.Printf("Failed to load corresponding response struct in method %v\n", m.Name)
				os.Exit(-1)
			}

			contract.Methods = append(contract.Methods, impl.NewMethod(pbPackage, m, request, response, goBindingFile, contract.StructName))

		}
		contract.Write(filepath, goType+"_server")
	}
}

var (
	predefinedStructs = map[string]struct{}{
		"TransactionResp": struct{}{},
		"Empty":           struct{}{},
		"TransactionReq":  struct{}{},
	}
)

func findGoStruct(name string, goFile *parser.GoFile) *parser.GoStruct {
	// retrun empty struct to handle default types
	_, ok := predefinedStructs[name]
	if ok {
		return &parser.GoStruct{
			Name: name,
		}
	}
	for _, s := range goFile.Structs {
		if name == s.Name {
			return s
		}
	}
	return nil
}

