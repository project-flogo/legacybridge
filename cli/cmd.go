package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/project-flogo/cli/common"
	"github.com/project-flogo/cli/util"
	"github.com/project-flogo/legacybridge/config"
	"github.com/spf13/cobra"
)

func init() {
	pluginUpgrade.Flags().StringVarP(&outFile, "output", "o", "", "specify output flogo.json file")
	legacyCmd.AddCommand(pluginUpgrade)
	common.RegisterPlugin(legacyCmd)
}

var outFile string

var legacyCmd = &cobra.Command{
	Use:   "legacy",
	Short: "work with legacy flogo",
	Long:  "Work with legacy flogo apps",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		//common.SetVerbose(verbose)
	},
}

var pluginUpgrade = &cobra.Command{
	Use:   "upgrade [flogo.json]",
	Short: "upgrade flogo.json",
	Long:  "Upgrades the flogo.json file",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) > 0 && args[0] != "" {

			var appJson string
			var err error

			flogoJsonPath := args[0]
			if util.IsRemote(flogoJsonPath) {

				appJson, err = util.LoadRemoteFile(flogoJsonPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: unable to load remote app file '%s' - %s", flogoJsonPath, err.Error())
					os.Exit(1)
				}
			} else {
				appJson, err = util.LoadLocalFile(flogoJsonPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: unable to load app file '%s' - %s", flogoJsonPath, err.Error())
					os.Exit(1)
				}
			}

			newJson, err := config.ConvertLegacyJson(appJson)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: converting legacy flogo.json: %s", err.Error())
				os.Exit(1)
			}

			currentDir, err := os.Getwd()
			if err != nil {
				currentDir = "."
			}

			newJsonFile := filepath.Join(currentDir, "flogo.json")

			if outFile != "" {
				newJsonFile = outFile
			}

			err = ioutil.WriteFile(newJsonFile, []byte(newJson), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: unable to write new json file: %s", err.Error())
				os.Exit(1)
			}

		} else {
			//upgrade existing flogo.json

			project := common.CurrentProject()
			workingDir := project.Dir()

			flogoJsonPath := filepath.Join(workingDir, "flogo.json")
			appJson, err := util.LoadLocalFile(flogoJsonPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: unable to load app file '%s' - %s", flogoJsonPath, err.Error())
				os.Exit(1)
			}

			newJson, err := config.ConvertLegacyJson(appJson)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: converting legacy flogo.json: %s", err.Error())
				os.Exit(1)
			}

			newJsonFile := flogoJsonPath

			if outFile != "" {
				newJsonFile = outFile
			} else {
				//rename existing flogo.json
				os.Rename(flogoJsonPath, filepath.Join(workingDir,"flogo.json.bak"))
			}

			err = ioutil.WriteFile(newJsonFile, []byte(newJson), 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: unable to write new json file: %s", err.Error())
				os.Exit(1)
			}
		}
	},
}
