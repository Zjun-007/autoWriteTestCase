package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/webnode/autoWriteTestCase/internal/config"
	"github.com/webnode/autoWriteTestCase/internal/pipeline"
)

var (
	version = "0.1.0"
	cfgPath string
)

func main() {
	root := &cobra.Command{
		Use:   "awtc",
		Short: "autoWriteTestCase — generate test cases from requirements",
	}
	root.PersistentFlags().StringVar(&cfgPath, "config", "configs/default.yaml", "config file path")

	root.AddCommand(newVersionCmd(), newGenerateCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	}
}

func newGenerateCmd() *cobra.Command {
	var (
		input   string
		outDir  string
		formats string
	)
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate test cases from a Markdown requirement",
		RunE: func(cmd *cobra.Command, args []string) error {
			if input == "" {
				return fmt.Errorf("--input is required")
			}
			cfg, err := config.Load(cfgPath)
			if err != nil {
				return err
			}

			log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
			pipe, err := pipeline.New(cfg, log)
			if err != nil {
				return err
			}

			var formatList []string
			if formats != "" {
				formatList = splitCSV(formats)
			}

			result, err := pipe.Run(context.Background(), pipeline.Options{
				Input:   input,
				OutDir:  outDir,
				Formats: formatList,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stdout, "generated %d cases from %d criteria → %s\n",
				len(result.Cases), len(result.Graph.Criteria), outDir)
			if !result.Report.OK {
				fmt.Fprintf(os.Stderr, "validation has %d error(s)\n", len(result.Report.Errors))
				for _, e := range result.Report.Errors {
					fmt.Fprintf(os.Stderr, "  - %s\n", e)
				}
			}
			for _, w := range result.Report.Warnings {
				fmt.Fprintf(os.Stderr, "warning: %s\n", w)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&input, "input", "i", "", "requirement markdown path")
	cmd.Flags().StringVarP(&outDir, "out", "o", "./out", "output directory")
	cmd.Flags().StringVarP(&formats, "formats", "f", "", "comma-separated formats: json,excel,markdown (default from config)")
	_ = cmd.MarkFlagRequired("input")
	return cmd
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
