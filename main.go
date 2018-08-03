package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

const name = "helm-helper"
const version = "0.1.0"
const description = "compare chart version, up version and create helm deploy"

type RepoMetadata struct {
	Entries map[string][]ChartRepoMetadata `yaml:"entries"`
}

type ChartRepoMetadata struct {
	Version string `yaml:"version"`
}

type Values struct {
	Metadata struct {
		Version   string        `yaml:"version"`
		Name      string        `yaml:"name"`
		Chart     string        `yaml:"chart"`
		Namespace string        `yaml:"namespace"`
		Vault     []VaultStruct `yaml:"vault"`
	} `yaml:"_metadata"`
}

type VaultStruct struct {
	Key   string `yaml:"key"`
	Field string `yaml:"field"`
	Path  string `yaml:"path"`
}

type builder interface {
	setVersion(string) builder
	build() HelmCommand
	setName(string) builder
	setChart(string) builder
	setNamespace(string) builder
	setValues(string) builder
	setVault([]VaultStruct) builder
}

type Helmbuilder struct {
	buffer strings.Builder
}

type HelmCommand struct {
	Cmd string
}

func errorBuild(field string) {
	var example = ` Example:
_metadata:
  chart: ritmx/prometheus
  name: prom
  namespace: log
  vault:
  - field: google_credentials_file
    key: gcssa
    path: secret/gcp/sandbox/thanos-sa
  version: 0.4.4
`
	log.Printf("field _metadata.%s empty", field)
	log.Fatal(example)

}

func (hc *Helmbuilder) setVersion(version string) builder {
	if version == "" {
		errorBuild("version")
	}
	fmt.Fprintf(&hc.buffer, " --version %s", version)
	return hc
}

func (hc *Helmbuilder) setNamespace(ns string) builder {
	if ns == "" {
		errorBuild("namespace")
	}
	fmt.Fprintf(&hc.buffer, " --namespace %s", ns)
	return hc
}

func (hc *Helmbuilder) setValues(values string) builder {
	fmt.Fprintf(&hc.buffer, " --values %s", values)
	return hc
}

func (hc *Helmbuilder) setChart(chart string) builder {
	if chart == "" {
		errorBuild("chart")
	}
	fmt.Fprintf(&hc.buffer, " %s", chart)
	return hc
}

func (hc *Helmbuilder) setName(name string) builder {
	if name == "" {
		errorBuild("name")
	}
	fmt.Fprintf(&hc.buffer, " %s", name)
	return hc
}

func (hc *Helmbuilder) setVault(vault []VaultStruct) builder {
	if len(vault) > 0 {
		for _, v := range vault {
			fmt.Fprintf(&hc.buffer, " --set %s=$(shell vault read -field %s %s)", v.Key, v.Field, v.Path)
		}

	}
	return hc
}

func (hc *Helmbuilder) build() HelmCommand {
	return HelmCommand{hc.buffer.String()}
}

func new(init string) builder {
	builder := Helmbuilder{}
	builder.buffer.WriteString(init)
	return &builder
}

func readValues(filepath string) Values {
	file, e := ioutil.ReadFile(filepath)
	if e != nil {
		log.Fatalf("File error: %v\n", e)
	}
	values := Values{}
	yaml.Unmarshal(file, &values)
	return values
}

// DL metadata from helm repo and compare version between helm repo and local values
func compareVersion(version string, chartName string, url string) {
	resp, err := http.Get(fmt.Sprintf("%s/index.yaml", url))
	if err != nil {
		log.Fatalf("Download metadata chart error: %s", err)
	}
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		log.Fatalf("Read metadata chart error: %s", e)
	}

	charts := RepoMetadata{}
	yaml.Unmarshal(body, &charts)

	metadata, ok := charts.Entries[strings.Split(chartName, "/")[1]]
	if !ok {
		log.Fatalf("%s doesnt exist", chartName)
	}
	c, _ := semver.NewConstraint(fmt.Sprintf("= %s", metadata[0].Version))
	v, _ := semver.NewVersion(version)
	b := c.Check(v)
	if !b {
		log.Fatalf("local version: %s repo version: %s", version, metadata[0].Version)
	}
}

func Check_version(values string, url string) {
	if values == "" {
		log.Fatalf("--values or -f arg is missed")
	}
	if url == "" {
		log.Fatalf("--url is missed")
	}
	valuesStruct := readValues(values)
	compareVersion(valuesStruct.Metadata.Version, valuesStruct.Metadata.Chart, url)
}

func Command(values string) {
	if values == "" {
		log.Fatalf("--values or -f arg is missed")
	}
	valuesStruct := readValues(values)
	builder := new("helm upgrade --install")
	cmd := builder.
		setChart(valuesStruct.Metadata.Chart).
		setName(valuesStruct.Metadata.Name).
		setVersion(valuesStruct.Metadata.Version).
		setNamespace(valuesStruct.Metadata.Namespace).
		setValues(values).
		setVault(valuesStruct.Metadata.Vault).
		build()
	fmt.Print(cmd.Cmd)
}

func main() {
	app := cli.NewApp()
	app.Name = name
	app.Usage = description
	app.Version = version
	app.ArgsUsage = "--values file"
	app.Commands = []cli.Command{
		{
			Name:  "check_version",
			Usage: "compare version between deployement and latest chart",
			Action: func(c *cli.Context) {
				Check_version(c.String("values"), c.String("url"))
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "values, f",
					Usage: "helm values `file`",
				},
				cli.StringFlag{
					Name:   "url,u",
					Usage:  "url chart repo",
					EnvVar: "HELM_URL",
					Value:  "https://kubernetes-charts.storage.googleapis.com",
				},
			},
		},
		{
			Name:  "command",
			Usage: "create helm command with _metadata block in values file",
			Action: func(c *cli.Context) {
				Command(c.String("values"))
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "values, f",
					Usage: "helm values `file`",
				},
			},
		},
	}
	app.Run(os.Args)
}
