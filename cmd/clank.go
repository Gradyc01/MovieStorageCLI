package cmd

import (
	"bufio"
	"context"
	"fmt"
	"movie-tracker/internal/environment"
	"movie-tracker/internal/gemini"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/genai"
)

// clankCmd defines "movie-tracker clank <str>". This command can be used to run all the other commands available
// inside the CLI whilst using Natural Language instead.
var clankCmd = &cobra.Command{
	Use:   "clank <string>",
	Short: "Asking AI to process your requests by using natural language",
	RunE: func(command *cobra.Command, args []string) error {

		ctx := context.Background()

		apiKey, err := environment.GetVariable("GEMINI_API_KEY")
		if err != nil {
			return err
		}
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:  apiKey,
			Backend: genai.BackendGeminiAPI,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create Gemini client:", err)
			os.Exit(1)
		}

		executor := gemini.NewExecutor(store)
		session := gemini.NewSession(client, executor)

		fmt.Println("Movie Storage assistant - type a request, or 'exit' to quit.")
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("\n> ")
			line, err := reader.ReadString('\n')
			if err != nil {
				return err
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if line == "exit" || line == "quit" {
				return nil
			}

			reply, err := session.Handle(ctx, line)
			if err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
				continue
			}
			fmt.Println(reply)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clankCmd)
}
