package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/Emptyless/yamlfmt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func main() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openapi-fmt",
		Short: "opinionated formatter of openapi.yaml files",
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath, err := cmd.Flags().GetString("file")
			if err != nil {
				return err
			}

			b, err := os.ReadFile(filePath)
			if err != nil {
				return err
			}

			node := new(yaml.Node)
			err = yaml.Unmarshal(b, node)
			if err != nil {
				return err
			}

			rules := yamlfmt.DefaultOpenAPIRules()
			// add alphabetical rules
			alphabeticalRules, err := cmd.Flags().GetStringArray("alphabetical")
			if err != nil {
				return err
			}
			for _, rule := range alphabeticalRules {
				rules = append(rules, yamlfmt.NewRule(rule, yamlfmt.StringOrderingFn))
			}

			// add simple rules
			simpleRules, err := cmd.Flags().GetStringArray("simple")
			if err != nil {
				return err
			}
			for _, rule := range simpleRules {
				splitted := strings.SplitN(rule, "=", 2)
				if len(splitted) != 2 {
					return fmt.Errorf("invalid rule format: %q, should be key=value,value2,...,valueN", rule)
				}

				rules = append(rules, yamlfmt.NewRule(splitted[0], yamlfmt.NewSimpleOrdering(strings.Split(splitted[1], ",")...)))
			}

			// validate rules
			err = yamlfmt.Validate(rules)
			if err != nil {
				return err
			}

			// lint document
			yamlfmt.Lint(node, rules)

			writer := new(bytes.Buffer)
			encoder := yaml.NewEncoder(writer)
			encoder.SetIndent(2)
			err = encoder.Encode(node)
			if err != nil {
				return err
			}

			outputPath, err := cmd.Flags().GetString("output")
			if err != nil && outputPath != "" {
				return os.WriteFile(outputPath, writer.Bytes(), 0644)
			}

			_, err = cmd.OutOrStdout().Write(writer.Bytes())
			return err
		},
	}

	cmd.Flags().StringP("file", "f", "", "path to openapi.yaml file")
	_ = cmd.MarkFlagRequired("file")
	cmd.Flags().StringP("output", "o", "", "path to output file")
	cmd.Flags().StringArrayP("alphabetical", "", []string{}, "path to node to sort alphabetically (e.g. '$.key')")
	cmd.Flags().StringArrayP("simple", "", []string{}, "path=keys to node to sort (e.g. path = '$.key') with comma separated list of keys")

	return cmd
}
