package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"
)

func NewCmdQuery() *cobra.Command {
	var jqFlags struct {
		NullInput bool
		RawInput  bool
		RawOutput bool
		Slurp     bool
		Arg       []string
		ArgJSON   []string
	}

	queryCmd := &cobra.Command{
		Use:     "query <query>",
		Short:   "Transform or generate JSON using a jq query",
		GroupID: "core",
		Args:    cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2)),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			vars := make([]string, 0)
			values := make([]interface{}, 0)
			for _, arg := range jqFlags.Arg {
				tokens := strings.SplitN(arg, "=", 2)
				if len(tokens) != 2 {
					log.Fatalln("invalid argument:", arg)
				}
				vars = append(vars, fmt.Sprintf("$%s", tokens[0]))
				values = append(values, tokens[1])
			}

			for _, arg := range jqFlags.ArgJSON {
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

			query, err := gojq.Parse(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			code, err := gojq.Compile(query, gojq.WithVariables(vars))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			var inputFile *os.File
			if len(args) == 2 {
				inputFile, err = os.Open(args[1])
				if err != nil {
					return fmt.Errorf("could not open file: %w", err)
				}
			} else {
				inputFile = os.Stdin
			}
			var inputs []interface{}
			if jqFlags.NullInput {
				inputs = append(inputs, nil)
			} else if jqFlags.RawInput {
				reader := bufio.NewReader(inputFile)
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						break
					}
					inputs = append(inputs, strings.TrimRight(line, "\n"))
				}
			} else {
				decoder := json.NewDecoder(inputFile)
				for {
					var v interface{}
					if err := decoder.Decode(&v); err != nil {
						break
					}
					inputs = append(inputs, v)
				}
			}

			var outputs []gojq.Iter
			if jqFlags.Slurp {
				if jqFlags.RawInput {
					input := strings.Builder{}
					for _, v := range inputs {
						input.WriteString(v.(string))
						input.WriteString("\n")
					}
					outputs = append(outputs, code.Run(input.String(), values...))
				} else {
					outputs = append(outputs, code.Run(inputs, values...))
				}
			} else {
				outputs = make([]gojq.Iter, len(inputs))
				for i, input := range inputs {
					outputs[i] = code.Run(input, values...)
				}
			}

			encoder := json.NewEncoder(os.Stdout)
			for _, output := range outputs {
				for {
					v, ok := output.Next()
					if !ok {
						break
					}
					if err, ok := v.(error); ok {
						return err
					}
					if jqFlags.RawOutput {
						fmt.Println(v)
						continue
					}
					err := encoder.Encode(v)
					if err != nil {
						return err
					}
				}
			}
			return nil
		},
	}

	queryCmd.Flags().BoolVarP(&jqFlags.NullInput, "null-input", "n", false, "use null as input value")
	queryCmd.Flags().BoolVarP(&jqFlags.RawInput, "raw-input", "R", false, "read input as raw strings")
	queryCmd.Flags().BoolVarP(&jqFlags.RawOutput, "raw-output", "r", false, "output raw strings, not JSON texts")
	queryCmd.Flags().BoolVarP(&jqFlags.Slurp, "slurp", "s", false, "read all inputs into an array")
	queryCmd.Flags().StringArrayVar(&jqFlags.Arg, "arg", []string{}, "add string variable in the form of name=value")
	queryCmd.Flags().StringArrayVar(&jqFlags.ArgJSON, "argjson", []string{}, "add JSON variable in the form of name=value")

	return queryCmd
}
