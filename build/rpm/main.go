package main

import (
	"bytes"
	"fmt"
	"github.com/ground-x/go-gxplatform/params"
	"github.com/urfave/cli"
	"os"
	"text/template"
)

type RpmSpec struct {
	BuildNumber int
	Version     string
	Name        string
}

func (r RpmSpec) String() string {
	tmpl, err := template.New("rpmspec").Parse(rpmSpecTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template, %v", err)
		return ""
	}

	result := new(bytes.Buffer)
	err = tmpl.Execute(result, r)
	if err != nil {
		fmt.Printf("Failed to render template, %v", err)
		return ""
	}
	return result.String()
}

func main() {
	app := cli.NewApp()
	app.Name = "klaytn_rpmtool"
	app.Version = "0.1"
	app.Commands = []cli.Command{
		{
			Name:    "gen_spec",
			Aliases: []string{"a"},
			Usage:   "generate rpm sepc file",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "devel",
					Usage: "generate spec for devel version",
				},
				cli.IntFlag{
					Name:  "build_num",
					Usage: "build number",
				},
			},
			Action: genspec,
		},
		{
			Name:    "version",
			Aliases: []string{"v"},
			Usage:   "return klaytn version",
			Action: func(c *cli.Context) error {
				fmt.Print(params.Version)
				return nil
			},
		},
		{
			Name:    "release_num",
			Aliases: []string{"r"},
			Usage:   "return klaytn release number",
			Action: func(c *cli.Context) error {
				fmt.Print(params.ReleaseNum)
				return nil
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func genspec(c *cli.Context) error {
	rpmSpec := new(RpmSpec)
	if c.Bool("devel") {
		buildNum := c.Int("build_num")
		if buildNum == 0 {
			fmt.Println("BuildNumber should be set")
			os.Exit(1)
		}
		rpmSpec.BuildNumber = buildNum
		rpmSpec.Name = "klaytn-devel"
	} else {
		rpmSpec.BuildNumber = params.ReleaseNum
		rpmSpec.Name = "klaytn"
	}
	rpmSpec.Version = params.Version
	fmt.Println(rpmSpec)
	return nil
}

var rpmSpecTemplate = `Name:               {{ .Name }}
Version:            {{ .Version }}
Release:            {{ .BuildNumber }}%{?dist}
Summary:            the Klaytn package

Group:              Application/blockchain
License:            GNU
URL:                http://www.klaytn.io
Source0:            %{name}-%{version}.tar.gz
BuildRoot:          %(mktemp -ud %{_tmppath}/%{name}-%{version}-%{release}-XXXXXX)

%description
 The Klaytn blockchain platform

%prep
%setup -q

%build
make klay

%install
mkdir -p $RPM_BUILD_ROOT/usr/local/bin
mkdir -p $RPM_BUILD_ROOT/etc/klay/conf
mkdir -p $RPM_BUILD_ROOT/etc/init.d/
mkdir -p $RPM_BUILD_ROOT/var/log/klay

cp build/bin/klay $RPM_BUILD_ROOT/usr/local/bin/klay
cp build/rpm/etc/init.d/klay $RPM_BUILD_ROOT/etc/init.d/klay
cp build/rpm/etc/klay/conf/klay.conf $RPM_BUILD_ROOT/etc/klay/conf/klay.conf

%files
%attr(754, -, -) /usr/local/bin/klay
%attr(644, -, -) /etc/klay/conf/klay.conf
%attr(754, -, -) /etc/init.d/klay
`
