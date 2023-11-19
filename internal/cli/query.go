package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func NewCmdQuery() *cobra.Command {
	var flags struct {
		NullInput     bool
		YamlInput     bool
		YamlOutput    bool
		RawInput      bool
		RawOutput     bool
		InPlace       bool
		Slurp         bool
		CompactOutput bool
		Arg           []string
		ArgJSON       []string
	}

	queryCmd := &cobra.Command{
		Use:     "query [query] [file]",
		Short:   "Transform or generate JSON using a jq query",
		Args:    cobra.MatchAll(cobra.MaximumNArgs(2)),
		GroupID: CommandGroupCore,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.InPlace && len(args) < 2 {
				return fmt.Errorf("a file argument is required with the --in-place flags")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			vars := make([]string, 0)
			values := make([]interface{}, 0)
			for _, arg := range flags.Arg {
				tokens := strings.SplitN(arg, "=", 2)
				if len(tokens) != 2 {
					log.Fatalln("invalid argument:", arg)
				}
				vars = append(vars, fmt.Sprintf("$%s", tokens[0]))
				values = append(values, tokens[1])
			}

			for _, arg := range flags.ArgJSON {
				tokens := strings.SplitN(arg, "=", 2)
				if len(tokens) != 2 {
					log.Fatalln("invalid argument:", arg)
				}
				vars = append(vars, fmt.Sprintf("$%s", tokens[0]))
				var value interface{}
				err = json.Unmarshal([]byte(tokens[1]), &value)
				if err != nil {
					log.Fatalln("invalid JSON:", arg)
				}
				values = append(values, value)
			}

			var rawQuery string
			if len(args) > 0 {
				rawQuery = args[0]
			} else {
				rawQuery = "."
			}

			query, err := gojq.Parse(rawQuery)
			if err != nil {
				return fmt.Errorf("could not parse query: %s", err)
			}
			code, err := gojq.Compile(query, gojq.WithVariables(vars))
			if err != nil {
				return fmt.Errorf("could not compile query: %s", err)
			}

			var inputFile *os.File
			if len(args) == 2 {
				inputFile, err = os.Open(args[1])
				if err != nil {
					return fmt.Errorf("could not open file: %s", err)
				}
			} else {
				inputFile = os.Stdin
			}
			var inputs []interface{}
			if flags.NullInput {
				inputs = append(inputs, nil)
			} else if flags.RawInput {
				reader := bufio.NewReader(inputFile)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						break
					}
					inputs = append(inputs, strings.TrimRight(line, "\n"))
				}
			} else {
				var decoder interface {
					Decode(any) error
				}

				if flags.YamlInput {
					decoder = yaml.NewDecoder(inputFile)
				} else {
					decoder = json.NewDecoder(inputFile)
				}

				for {
					var v interface{}
					if err := decoder.Decode(&v); err != nil {
						break
					}
					inputs = append(inputs, v)
				}
			}
			if err := inputFile.Close(); err != nil {
				return err
			}

			outputs := make([]gojq.Iter, 0)
			if flags.Slurp {
				if flags.RawInput {
					input := strings.Builder{}
					for _, v := range inputs {
						input.WriteString(v.(string))
						input.WriteString("\n")
					}
					outputs = append(outputs, code.Run(input.String(), values...))
				} else if flags.NullInput {
					outputs = append(outputs, code.Run(nil, values...))
				} else {
					outputs = append(outputs, code.Run(inputs, values...))
				}
			} else {
				outputs = make([]gojq.Iter, len(inputs))
				for i, input := range inputs {
					outputs[i] = code.Run(input, values...)
				}
			}

			outputFile := os.Stdout
			if flags.InPlace {
				f, err := os.OpenFile(args[1], os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}

				outputFile = f
			}

			var encoder interface {
				Encode(any) error
			}

			if flags.YamlOutput {
				encoder = yaml.NewEncoder(outputFile)
			} else {
				jsonEncoder := json.NewEncoder(outputFile)
				if !flags.CompactOutput {
					jsonEncoder.SetIndent("", "    ")
				}

				encoder = jsonEncoder
			}
			for _, output := range outputs {
				for {
					v, ok := output.Next()
					if !ok {
						break
					}
					if err, ok := v.(error); ok {
						return fmt.Errorf("could not run query: %s", err)
					}
					if flags.RawOutput {
						if s, ok := v.(string); ok {
							fmt.Println(s)
							continue
						}
					}

					// go encode empty array as null, we want an empty array
					if s, ok := v.([]interface{}); ok && len(s) == 0 {
						v = make([]interface{}, 0)
					}

					err := encoder.Encode(v)
					if err != nil {
						return fmt.Errorf("could not encode output: %s", err)
					}
				}
			}
			return nil
		},
	}

	queryCmd.Flags().BoolVarP(&flags.NullInput, "null-input", "n", false, "use null as input value")
	queryCmd.Flags().BoolVarP(&flags.RawInput, "raw-input", "R", false, "read input as raw strings")
	queryCmd.Flags().BoolVarP(&flags.RawOutput, "raw-output", "r", false, "output raw strings, not JSON texts")
	queryCmd.Flags().BoolVarP(&flags.Slurp, "slurp", "s", false, "read all inputs into an array")
	queryCmd.Flags().BoolVar(&flags.YamlInput, "yaml-input", false, "read input as YAML format")
	queryCmd.Flags().BoolVarP(&flags.InPlace, "in-place", "i", false, "read and write to the same file")
	queryCmd.Flags().BoolVar(&flags.YamlOutput, "yaml-output", false, "output as YAML")
	queryCmd.Flags().StringArrayVar(&flags.Arg, "arg", []string{}, "add string variable in the form of name=value")
	queryCmd.Flags().StringArrayVar(&flags.ArgJSON, "argjson", []string{}, "add JSON variable in the form of name=value")
	queryCmd.Flags().BoolVarP(&flags.CompactOutput, "compact-output", "c", false, "output without pretty-printing")

	return queryCmd
}
