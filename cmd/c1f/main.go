package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/c1f/c1f/pkg/api"
	"github.com/c1f/c1f/pkg/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func main() {
	var workflowName string
	var instanceID string
	var debug bool

	rootCmd := &cobra.Command{
		Use:   "c1f",
		Short: "Cloudflare Workflows CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no subcommand is specified, run the TUI
			apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
			accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

			if apiToken == "" {
				return fmt.Errorf("CLOUDFLARE_API_TOKEN is required")
			}
			if accountID == "" {
				return fmt.Errorf("CLOUDFLARE_ACCOUNT_ID is required")
			}

			client := api.NewClient(apiToken, accountID)
			m := ui.NewRootModel(client)
			p := tea.NewProgram(m, tea.WithAltScreen())
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("error running TUI: %w", err)
			}
			return nil
		},
	}

	describeCmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe a workflow instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			apiToken := os.Getenv("CLOUDFLARE_API_TOKEN")
			accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")

			if apiToken == "" {
				return fmt.Errorf("CLOUDFLARE_API_TOKEN is required")
			}
			if accountID == "" {
				return fmt.Errorf("CLOUDFLARE_ACCOUNT_ID is required")
			}
			if workflowName == "" {
				return fmt.Errorf("--workflow is required")
			}
			if instanceID == "" {
				return fmt.Errorf("--instance is required")
			}

			client := api.NewClient(apiToken, accountID)
			client.Debug = debug
			instance, err := client.GetWorkflowInstance(context.Background(), workflowName, instanceID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if instance.Status == "running" {
				_, progressStr := instance.CalculateProgress()
				fmt.Printf("Instance ID: %s\nStatus: %s\nProgress: %s\nSteps: %d\n", 
					instance.ID, instance.Status, progressStr, len(instance.Steps))
			} else {
				fmt.Printf("Instance ID: %s\nStatus: %s\nSteps: %d\n", 
					instance.ID, instance.Status, len(instance.Steps))
			}

			out, err := json.MarshalIndent(instance, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal instance: %w", err)
			}
			fmt.Println(string(out))
			return nil
		},
	}

	describeCmd.Flags().StringVar(&workflowName, "workflow", "", "The name of the workflow")
	describeCmd.Flags().StringVar(&instanceID, "instance", "", "The ID of the workflow instance")
	describeCmd.Flags().BoolVar(&debug, "debug", false, "Enable raw request/response logging to stderr")

	rootCmd.AddCommand(describeCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
