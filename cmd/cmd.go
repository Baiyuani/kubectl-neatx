/*
Copyright © 2019 Itay Shakury @itaysk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"regexp"
	s "strings"
	"unicode"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

var outputFormat *string
var inputFile *string
var namespace *string
var exportOutDir *string
var allNamespaces *bool

func init() {
	outputFormat = rootCmd.PersistentFlags().StringP("output", "o", "yaml", "output format: yaml or json")
	inputFile = rootCmd.Flags().StringP("file", "f", "-", "file path to neat, or - to read from stdin")
	namespace = exportCmd.Flags().StringP("namespace", "n", "-", "namespace")
	// kindListFromFile = exportCmd.Flags().StringP("list-file", "l", "-", "file path to kind list from file")
	exportOutDir = exportCmd.Flags().StringP("dest-dir", "d", "-", "export file to directory")
	allNamespaces = exportCmd.Flags().BoolP("all-namespaces", "A", false, "export all namespaces")
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.MarkFlagFilename("file")
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(exportCmd)
}

// Execute is the entry point for the command package
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use: "kubectl-neat",
	Example: `kubectl get pod mypod -o yaml | kubectl neat
kubectl get pod mypod -oyaml | kubectl neat -o json
kubectl neat -f - <./my-pod.json
kubectl neat -f ./my-pod.json
kubectl neat -f ./my-pod.json --output yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var in, out []byte
		var err error
		if *inputFile == "-" {
			stdin := cmd.InOrStdin()
			in, err = io.ReadAll(stdin)
			if err != nil {
				return err
			}
		} else {
			in, err = os.ReadFile(*inputFile)
			if err != nil {
				return err
			}
		}
		outFormat := *outputFormat
		if !cmd.Flag("output").Changed {
			outFormat = "same"
		}
		out, err = NeatYAMLOrJSON(in, outFormat)
		if err != nil {
			return err
		}
		cmd.Print(string(out))
		return nil
	},
}

var kubectl string = "kubectl"

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Print specific resource manifest",
	Example: `kubectl neat get -- pod mypod -oyaml
kubectl neat get -- svc -n default myservice --output json`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true}, //don't try to validate kubectl get's flags
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := get(cmd, args)
		if err != nil {
			return err
		}
		cmd.Println(out)
		return nil
	},
}

// populated by goreleaser
var (
	Version = "v0.0.0+unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print kubectl-neat version",
	Long:  "Print the version of kubectl-neat",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kubectl-neat version: %s\n", Version)
	},
}

func isJSON(s []byte) bool {
	return bytes.HasPrefix(bytes.TrimLeftFunc(s, unicode.IsSpace), []byte{'{'})
}

// NeatYAMLOrJSON converts 'in' to json if needed, invokes neat, and converts back if needed according the the outputFormat argument: yaml/json/same
func NeatYAMLOrJSON(in []byte, outputFormat string) (out []byte, err error) {
	var injson, outjson string
	itsYaml := !isJSON(in)
	if itsYaml {
		injsonbytes, err := yaml.YAMLToJSON(in)
		if err != nil {
			return nil, fmt.Errorf("error converting from yaml to json : %v", err)
		}
		injson = string(injsonbytes)
	} else {
		injson = string(in)
	}

	outjson, err = Neat(injson)
	if err != nil {
		return nil, fmt.Errorf("error neating : %v", err)
	}

	if outputFormat == "yaml" || (outputFormat == "same" && itsYaml) {
		out, err = yaml.JSONToYAML([]byte(outjson))
		if err != nil {
			return nil, fmt.Errorf("error converting from json to yaml : %v", err)
		}
	} else {
		out = []byte(outjson)
	}
	return
}

func get(cmd *cobra.Command, args []string) (string, error) {
	var out []byte
	var err error
	//reset defaults
	//there are two output settings in this subcommand: kubectl get's and kubectl-neat's
	//any combination of those can be provided by using the output flag in either side of the --
	//the most efficient is kubectl: json, kubectl-neat: yaml
	//0--0->Y--J #choose what's best for us
	//0--Y->Y--Y #user did specify output in kubectl, so respect that
	//0--J->J--J #user did specify output in kubectl, so respect that
	//Y--0->Y--J #user doesn't care about kubectl so use json but convert back
	//J--0->J--J #user expects json so use it for foth
	//if the user specified both side we can't touch it

	//the desired kubectl get output is always json, unless it was explicitly set by the user to yaml in which case the arg is overriden when concatenating the args later
	cmdArgs := append([]string{"get", "-o", "json"}, args...)
	kubectlCmd := exec.Command(kubectl, cmdArgs...)
	kres, err := kubectlCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error invoking kubectl as %v %v", kubectlCmd.Args, err)
	}
	//handle the case of 0--J->J--J
	outFormat := *outputFormat
	kubeout := "yaml"
	for _, arg := range args {
		if arg == "json" || arg == "ojson" {
			outFormat = "json"
		}
	}
	if !cmd.Flag("output").Changed && kubeout == "json" {
		outFormat = "json"
	}
	out, err = NeatYAMLOrJSON(kres, outFormat)
	if err != nil {
		return "", err
	}
	// cmd.Println(string(out))
	return string(out), nil
}

func getAllNamespaces() ([]string, error) {
	var res []string
	cmd := exec.Command(kubectl, "get", "namespace", "-o", "name")
	cmdout, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	namespacesSLice := s.Split(string(cmdout), "\n")
	for _, namespace := range namespacesSLice[:len(namespacesSLice)-1] {
		res = append(res, s.Split(namespace, "/")[1])
	}
	return res, nil
}

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Batch export of specified resource lists",
	Example: `kubectl neat export -n default deploy sts ...`,
	// FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true}, //don't try to validate kubectl get's flags
	RunE: func(cmd *cobra.Command, args []string) error {
		// fmt.Println(args)
		// fmt.Printf("%s\n", *namespace)
		var namespacesList []string
		var err error

		//获取api-resources
		apiResourcesCmd := exec.Command(kubectl, "api-resources", "--no-headers")
		apiResourcesCmdOut, err := apiResourcesCmd.CombinedOutput()
		var apiResources []string
		if err != nil {
			apiResources = nil
			return fmt.Errorf("Error invoking kubectl as %v %v", apiResourcesCmd.Args, err)
		} else {
			apiResources = s.Split(string(apiResourcesCmdOut), "\n")[:len(s.Split(string(apiResourcesCmdOut), "\n"))-1]
		}

		//获取kind清单
		var kindList []string
		for _, arg := range args {
			kindList = append(kindList, s.Split(arg, ",")...)
		}

		//存储目录初始化
		var outDir string
		var clustrdDir = "Clusterd"
		if *exportOutDir == "-" {
			outDir = "manifests"
		} else {
			outDir = *exportOutDir
		}
		err = os.Mkdir(outDir, 0755)
		if err != nil {
			return err
		}
		err = os.Mkdir(path.Join(outDir, clustrdDir), 0755)
		if err != nil {
			return err
		}

		//获取命名空间slice
		if *allNamespaces {
			namespacesList, err = getAllNamespaces()
			if err != nil {
				return err
			}
		} else {
			namespacesList = append(namespacesList, s.Split(*namespace, ",")...)
		}

		//执行
		for _, kind := range kindList {
			var kindDir string
			//判断是否为clusterd的kind
			condition, _ := isClusetrdKind(kind, apiResources)
			if condition {
				kindDir = path.Join(outDir, clustrdDir, kind)
				err := os.Mkdir(kindDir, 0755)
				if err != nil {
					return err
				}
				_ = getManifest(cmd, kindDir, kind, "default")
				kindList = DeleteSlice3(kindList, kind)
			}
		}

		for _, ns := range namespacesList {
			nsDir := fmt.Sprintf("%s/%s", outDir, ns)
			err := os.Mkdir(nsDir, 0755)
			if err != nil {
				return err
			}

			for _, kind := range kindList {
				kindDir := path.Join(nsDir, kind)
				err := os.Mkdir(kindDir, 0755)
				if err != nil {
					return err
				}

				_ = getManifest(cmd, kindDir, kind, ns)

			}
		}
		return nil
	},
}

func isClusetrdKind(name string, cache []string) (bool, error) {
	if cache == nil {
		return false, nil
	}

	for _, api := range cache {
		apiSlice := s.Split(deleteExtraSpace(api), " ")
		if apiSlice[1] == name {
			fmt.Printf("DEBUG: %s, %s\n", name, apiSlice[3])
			if apiSlice[3] == "true" {
				return false, nil
			} else {
				return true, nil
			}
		}
	}

	for _, api := range cache {
		apiSlice := s.Split(deleteExtraSpace(api), " ")
		if s.Contains(apiSlice[0], name) {
			fmt.Printf("DEBUG: %s, %s\n", name, apiSlice[3])
			if apiSlice[3] == "true" {
				return false, nil
			} else {
				return true, nil
			}
		}
	}

	return false, nil
}

// 删除字符串中多余的空格，有多个空格时，仅保留一个空格
func deleteExtraSpace(data string) string {
	// 替换 tab 为空格
	s1 := s.Replace(data, "\t", " ", -1)

	// 正则表达式匹配两个及两个以上空格
	regstr := "\\s{2,}"
	reg, _ := regexp.Compile(regstr)

	// 将字符串复制到切片
	s2 := make([]byte, len(s1))
	copy(s2, s1)

	// 删除多余空格
	spcIndex := reg.FindStringIndex(string(s2))
	for len(spcIndex) > 0 {
		s2 = append(s2[:spcIndex[0]+1], s2[spcIndex[1]:]...)
		spcIndex = reg.FindStringIndex(string(s2))
	}

	return string(s2)
}

func getManifest(cmd *cobra.Command, kindDir string, kind string, ns string) error {

	//获取资源名字列表
	kubectlCmd := exec.Command(kubectl, "get", kind, "-n", ns, "-o", "name")
	kcmdRes, err := kubectlCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error invoking kubectl as %v %v", kubectlCmd.Args, err)
	}

	// fmt.Printf("%s\n", kcmdRes)

	resoucesSLice := s.Split(string(kcmdRes), "\n")
	for _, name := range resoucesSLice[:len(resoucesSLice)-1] {
		// fmt.Printf("debug: %s\n", string(name))
		out, err := get(cmd, []string{name, "-n", ns})
		if err != nil {
			return err
		}

		// fmt.Printf("%s\n", out)
		resourceFile := fmt.Sprintf("%s/%s.yaml", kindDir, s.Split(name, "/")[1])
		os.WriteFile(resourceFile, []byte(out), 0644)
	}
	return nil
}

func DeleteSlice3(a []string, elem string) []string {
	j := 0
	for _, v := range a {
		if v != elem {
			a[j] = v
			j++
		}
	}
	return a[:j]
}
