package main

import (
        "fmt"
        "os"
        "strings"

        "github.com/spf13/cobra"
        "github.com/spf13/pflag"
        "github.com/takutakahashi/operation-mcp/pkg/config"
        "github.com/takutakahashi/operation-mcp/pkg/tool"
)

var (
        configPath string
        cfg        *config.Config
        toolMgr    *tool.Manager
)

func main() {
        rootCmd := &cobra.Command{
                Use:   "operations",
                Short: "Operations CLI tool",
                Long:  "A CLI tool for executing operations defined in a configuration file",
                PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
                        var err error
                        cfg, err = config.LoadConfig(configPath)
                        if err != nil {
                                return fmt.Errorf("failed to load config: %w", err)
                        }

                        if err := cfg.Validate(); err != nil {
                                return fmt.Errorf("invalid configuration: %w", err)
                        }

                        toolMgr = tool.NewManager(cfg)
                        return nil
                },
        }

        rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "path to config file")

        // Add dynamic commands based on configuration
        if err := addDynamicCommands(rootCmd); err != nil {
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
        }

        if err := rootCmd.Execute(); err != nil {
                fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                os.Exit(1)
        }
}

func addDynamicCommands(rootCmd *cobra.Command) error {
        // Try to load config
        cfg, err := config.LoadConfig(configPath)
        if err != nil {
                // If config can't be loaded, just return without adding commands
                // They will be added in the PersistentPreRunE function
                return nil
        }

        // Create tool manager
        toolMgr = tool.NewManager(cfg)

        // Add commands for each tool
        for _, tool := range cfg.Tools {
                toolCmd := createToolCommand(tool)
                rootCmd.AddCommand(toolCmd)
        }

        return nil
}

func createToolCommand(tool config.Tool) *cobra.Command {
        toolCmd := &cobra.Command{
                Use:   tool.Name,
                Short: fmt.Sprintf("Execute %s command", tool.Name),
                Run: func(cmd *cobra.Command, args []string) {
                        // If no subtools, execute the tool directly
                        if len(tool.Subtools) == 0 {
                                paramValues := getParamValues(cmd, tool.Params)
                                if err := toolMgr.ExecuteTool(tool.Name, paramValues); err != nil {
                                        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                                        os.Exit(1)
                                }
                                return
                        }

                        // Otherwise, show help
                        cmd.Help()
                },
        }

        // Add flags for tool parameters
        addParamFlags(toolCmd, tool.Params)

        // Add subcommands for each subtool
        for _, subtool := range tool.Subtools {
                subtoolCmd := createSubtoolCommand(tool.Name, subtool)
                toolCmd.AddCommand(subtoolCmd)
        }

        return toolCmd
}

func createSubtoolCommand(parentName string, subtool config.Subtool) *cobra.Command {
        // Replace spaces with underscores in the name
        name := strings.ReplaceAll(subtool.Name, " ", "_")
        fullName := parentName + "_" + name

        subtoolCmd := &cobra.Command{
                Use:   name,
                Short: fmt.Sprintf("Execute %s command", fullName),
                Run: func(cmd *cobra.Command, args []string) {
                        // If no subtools, execute the subtool
                        if len(subtool.Subtools) == 0 {
                                paramValues := getParamValues(cmd, nil) // Get all flags
                                if err := toolMgr.ExecuteTool(fullName, paramValues); err != nil {
                                        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
                                        os.Exit(1)
                                }
                                return
                        }

                        // Otherwise, show help
                        cmd.Help()
                },
        }

        // Add flags for subtool parameters
        addParamFlags(subtoolCmd, subtool.Params)

        // Add subcommands for each nested subtool
        for _, nestedSubtool := range subtool.Subtools {
                nestedCmd := createSubtoolCommand(fullName, nestedSubtool)
                subtoolCmd.AddCommand(nestedCmd)
        }

        return subtoolCmd
}

func addParamFlags(cmd *cobra.Command, params config.Parameters) {
        for name, param := range params {
                switch param.Type {
                case "string":
                        cmd.Flags().String(name, "", param.Description)
                case "int", "number":
                        cmd.Flags().Int(name, 0, param.Description)
                case "bool", "boolean":
                        cmd.Flags().Bool(name, false, param.Description)
                default:
                        // Default to string for unknown types
                        cmd.Flags().String(name, "", param.Description)
                }

                if param.Required {
                        cmd.MarkFlagRequired(name)
                }
        }
}

func getParamValues(cmd *cobra.Command, params config.Parameters) map[string]string {
        result := make(map[string]string)

        // Get all flags
        cmd.Flags().Visit(func(flag *pflag.Flag) {
                result[flag.Name] = flag.Value.String()
        })

        return result
}