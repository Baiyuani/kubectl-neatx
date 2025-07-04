package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
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

//go:embed api-resources.txt
var folder embed.FS

func init() {
	outputFormat = rootCmd.PersistentFlags().StringP("output", "o", "yaml", "output format: yaml or json")
	inputFile = rootCmd.Flags().StringP("file", "f", "-", "file path to neat, or - to read from stdin")
	namespace = exportCmd.Flags().StringP("namespace", "n", "default", "namespace")
	// kindListFromFile = exportCmd.Flags().StringP("list-file", "l", "-", "file path to kind list from file")
	exportOutDir = exportCmd.Flags().StringP("dest-dir", "d", "manifests", "export file to directory")
	allNamespaces = exportCmd.Flags().BoolP("all-namespaces", "A", false, "export all namespaces")
	migrateCmd.Flags().String("source-context", "", "source cluster context name")
	migrateCmd.Flags().String("target-context", "", "target cluster context name")
	migrateCmd.MarkFlagRequired("source-context")
	migrateCmd.MarkFlagRequired("target-context")
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
	rootCmd.MarkFlagFilename("file")
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(migrateCmd)
}

// Execute is the entry point for the command package
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use: "kubectl-neatx",
	Example: `kubectl get pod mypod -o yaml | kubectl neatx
kubectl neatx -f - <./my-pod.json
kubectl neatx -f ./my-pod.json
kubectl neatx -f ./my-pod.json --output yaml`,
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
	Example: `kubectl neatx get -- pod mypod -oyaml
kubectl neatx get -- svc -n default myservice --output json`,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true}, //don't try to validate kubectl get's flags
	RunE: func(cmd *cobra.Command, args []string) error {
		outputFormat := outFmt(cmd, args)
		out, err := get(args, outputFormat)
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
	Short: "Print kubectl-neatx version",
	Long:  "Print the version of kubectl-neatx",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kubectl-neatx version: %s\n", Version)
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

func get(args []string, outFormat string) (string, error) {
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
		// return "", fmt.Errorf("error invoking kubectl as %v %v", cmdArgs, err)
		return "", err
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

func getapiResource() []string {
	//获取api-resources
	var apiResources []string

	content, _ := folder.ReadFile("api-resources.txt")

	apiResourcesCmd := exec.Command(kubectl, "api-resources", "--no-headers")
	apiResourcesCmdOut, err := apiResourcesCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("error invoking kubectl as %v %v, getting embedding api-resources..", apiResourcesCmd.Args, err)
		apiResources = s.Split(string(content), "\n")[:len(s.Split(string(content), "\n"))-1]
		// return fmt.Errorf("error invoking kubectl as %v %v", apiResourcesCmd.Args, err)
	} else {
		apiResources = s.Split(string(apiResourcesCmdOut), "\n")[:len(s.Split(string(apiResourcesCmdOut), "\n"))-1]
	}
	return apiResources
}

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate resources between clusters",
	Example: `kubectl neatx migrate --source-context=ctx1 --target-context=ctx2 deploy/myapp -n default
kubectl neatx migrate --source-context=ctx1 --target-context=ctx2 --all -n default`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sourceContext := cmd.Flag("source-context").Value.String()
		targetContext := cmd.Flag("target-context").Value.String()

		// Get resources from source cluster
		kubectlCmd := exec.Command(kubectl, "--context", sourceContext, "get", "-o", "json", args[0])
		kres, err := kubectlCmd.CombinedOutput()
		if err != nil {
			return err
		}

		// Neat the resource
		out, err := NeatYAMLOrJSON(kres, "json")
		if err != nil {
			return err
		}

		// Apply to target cluster
		applyCmd := exec.Command(kubectl, "--context", targetContext, "apply", "-f", "-")
		applyCmd.Stdin = bytes.NewReader(out)
		applyOut, err := applyCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("%s: %v", string(applyOut), err)
		}

		cmd.Println(string(applyOut))
		return nil
	},
}

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Batch export of specified resource manifests",
	Example: `kubectl neatx export -n default deploy,sts,svc ...`,
	// FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true}, //don't try to validate kubectl get's flags
	RunE: func(cmd *cobra.Command, args []string) error {
		var namespacesList []string
		var err error

		//获取kind清单
		var kindList []string
		for _, arg := range args {
			kindList = append(kindList, s.Split(arg, ",")...)
		}

		outputFormat := outFmt(cmd, args)

		//存储目录初始化
		var outDir string
		var clustrdDir = "Cluster"
		outDir = *exportOutDir
		// err = os.MkdirAll(outDir, 0755)
		// if err != nil {
		// 	return err
		// }
		// err = os.MkdirAll(path.Join(outDir, clustrdDir), 0755)
		// if err != nil {
		// 	return err
		// }

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
		apiResources := getapiResource()
		var namespacedKindList []string
		for _, kind := range kindList {
			var kindDir string
			//判断是否为Cluster的kind
			condition := isClusterKind(kind, apiResources)
			if condition {
				kindDir = path.Join(outDir, clustrdDir, kind)
				err := os.MkdirAll(kindDir, 0755)
				if err != nil {
					return err
				}
				getManifest(kindDir, kind, "default", outputFormat)

			} else {
				namespacedKindList = append(namespacedKindList, kind)
			}
		}

		for _, ns := range namespacesList {
			nsDir := fmt.Sprintf("%s/%s", outDir, ns)
			err := os.MkdirAll(nsDir, 0755)
			if err != nil {
				return err
			}

			for _, kind := range namespacedKindList {
				kindDir := path.Join(nsDir, kind)
				err := os.MkdirAll(kindDir, 0755)
				if err != nil {
					return err
				}

				getManifest(kindDir, kind, ns, outputFormat)

			}
		}
		return nil
	},
}

func isClusterKind(name string, cache []string) bool {
	if cache == nil {
		return false
	}

	for _, api := range cache {
		api = deleteExtraSpace(api)
		apiSlice := s.Split(api, " ")
		len := len(apiSlice)
		if len == 4 {
			if s.ToLower(apiSlice[3]) == name {
				if apiSlice[2] == "true" {
					return false
				} else {
					return true
				}
			}
			if apiSlice[0] == name {
				if apiSlice[2] == "true" {
					return false
				} else {
					return true
				}
			}
		}

		if len == 5 {
			if apiSlice[1] == name {
				if apiSlice[3] == "true" {
					return false
				} else {
					return true
				}
			}
			if s.ToLower(apiSlice[4]) == name {
				if apiSlice[3] == "true" {
					return false
				} else {
					return true
				}
			}
			if apiSlice[0] == name {
				if apiSlice[3] == "true" {
					return false
				} else {
					return true
				}
			}
		}

	}

	return false
}

func getManifest(kindDir string, kind string, ns string, outFmt string) {
	//获取资源名字列表
	kubectlCmd := exec.Command(kubectl, "get", kind, "-n", ns, "-o", "name")
	kcmdRes, err := kubectlCmd.CombinedOutput()
	if err != nil {
		// return fmt.Errorf("error invoking kubectl as %v %v", kubectlCmd.Args, err)
		fmt.Printf("%s", string(kcmdRes))
		os.Remove(kindDir)
	} else {
		resoucesSLice := s.Split(string(kcmdRes), "\n")
		for _, name := range resoucesSLice[:len(resoucesSLice)-1] {
			out, err := get([]string{name, "-n", ns}, outFmt)
			if err != nil {
				fmt.Printf("%v", err)
			} else {
				resourceFile := fmt.Sprintf("%s/%s.%s", kindDir, s.Split(name, "/")[1], outFmt)
				fmt.Println(resourceFile)
				os.WriteFile(resourceFile, []byte(out), 0644)
			}

		}
	}
}

func outFmt(cmd *cobra.Command, args []string) string {
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
	return outFormat
}
