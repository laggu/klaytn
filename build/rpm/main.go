package main

import (
	"bytes"
	"fmt"
	"github.com/ground-x/klaytn/params"
	"github.com/urfave/cli"
	"os"
	"strings"
	"text/template"
)

const (
	CN = "kcn"
	PN = "kpn"
	EN = "ken"
)

type NodeInfo struct {
	daemon  string
	summary string
}

var NODE_TYPE = map[string]NodeInfo{
	CN: {"kcnd", "kcnd is klaytn consensus node daemon"},
	PN: {"kpnd", "kpnd is klaytn proxy node daemon"},
	EN: {"kend", "kend is klaytn endpoint node daemon"},
}

type RpmSpec struct {
	BuildNumber int
	Version     string
	Name        string
	Summary     string
	MakeTarget  string
	ProgramName string // kcn, kpn, ken
	DaemonName  string // kcnd, kpnd, kend
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
	app.Version = "0.2"
	app.Commands = []cli.Command{
		{
			Name:    "gen_spec",
			Aliases: []string{"a"},
			Usage:   "generate rpm spec file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "node_type",
					Usage: "Klaytn node type (cn, pn, en)",
				},
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

	nodeType := c.String("node_type")
	if _, ok := NODE_TYPE[nodeType]; ok != true {
		return fmt.Errorf("node_type[\"%s\"] is not supported. Use --node_type [kcn, kpn, ken]", nodeType)
	}

	rpmSpec.ProgramName = strings.ToLower(nodeType)
	rpmSpec.DaemonName = NODE_TYPE[nodeType].daemon

	if c.Bool("devel") {
		buildNum := c.Int("build_num")
		if buildNum == 0 {
			fmt.Println("BuildNumber should be set")
			os.Exit(1)
		}
		rpmSpec.BuildNumber = buildNum
		rpmSpec.Name = NODE_TYPE[nodeType].daemon + "-devel"
	} else {
		rpmSpec.BuildNumber = params.ReleaseNum
		rpmSpec.Name = NODE_TYPE[nodeType].daemon
	}
	rpmSpec.Summary = NODE_TYPE[nodeType].summary
	rpmSpec.Version = params.Version
	fmt.Println(rpmSpec)
	return nil
}

var rpmSpecTemplate = `Name:               {{ .Name }}
Version:            {{ .Version }}
Release:            {{ .BuildNumber }}%{?dist}
Summary:            {{ .Summary }}

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
make {{ .ProgramName }}

%install
mkdir -p $RPM_BUILD_ROOT/usr/bin
mkdir -p $RPM_BUILD_ROOT/etc/{{ .DaemonName }}/conf
mkdir -p $RPM_BUILD_ROOT/etc/init.d
mkdir -p $RPM_BUILD_ROOT/var/log/{{ .DaemonName }}

cp build/bin/{{ .ProgramName }} $RPM_BUILD_ROOT/usr/bin/{{ .ProgramName }}
cp build/rpm/etc/init.d/{{ .DaemonName }} $RPM_BUILD_ROOT/etc/init.d/{{ .DaemonName }}
cp build/rpm/etc/{{ .DaemonName }}/conf/{{ .DaemonName }}.conf $RPM_BUILD_ROOT/etc/{{ .DaemonName }}/conf/{{ .DaemonName }}.conf

%files
%attr(755, -, -) /usr/bin/{{ .ProgramName }}
%attr(644, -, -) /etc/{{ .DaemonName }}/conf/{{ .DaemonName }}.conf
%attr(754, -, -) /etc/init.d/{{ .DaemonName }}
%config(noreplace) /etc/{{ .DaemonName }}/conf/{{ .DaemonName }}.conf
`
