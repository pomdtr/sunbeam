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

var jqCmd = &cobra.Command{
	Use:   "jq",
	Short: "Run a jq query",
	Run:   sunbeamJQ,
	Args:  cobra.ExactArgs(1),
}

var JQFlags struct {
	NullInput bool
	RawInput  bool
	Arg       []string
	ArgJSON   []string
}

func init() {
	rootCmd.AddCommand(jqCmd)
	jqCmd.Flags().BoolVarP(&JQFlags.NullInput, "null-input", "n", false, "use null as input value")
	jqCmd.Flags().BoolVarP(&JQFlags.RawInput, "raw-input", "R", false, "read input as raw strings")
	jqCmd.Flags().StringArrayVar(&JQFlags.Arg, "arg", []string{}, "add string variable in the form of name=value")
	jqCmd.Flags().StringArrayVar(&JQFlags.ArgJSON, "argjson", []string{}, "add JSON variable in the form of name=value")
}

func sunbeamJQ(cmd *cobra.Command, args []string) {
	var err error
	vars := make([]string, 0)
	values := make([]interface{}, 0)
	for _, arg := range JQFlags.Arg {
		tokens := strings.SplitN(arg, "=", 2)
		if len(tokens) != 2 {
			log.Fatalln("invalid argument:", arg)
		}
		vars = append(vars, fmt.Sprintf("$%s", tokens[0]))
		values = append(values, tokens[1])
	}

	for _, arg := range JQFlags.ArgJSON {
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

	var inputs []interface{}
	if JQFlags.NullInput {
		inputs = append(inputs, nil)
	} else if JQFlags.RawInput {
		reader := bufio.NewReader(os.Stdin)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			inputs = append(inputs, strings.TrimRight(line, "\n"))
		}
	} else {
		decoder := json.NewDecoder(os.Stdin)
		for {
			var v interface{}
			if err := decoder.Decode(&v); err != nil {
				break
			}
			inputs = append(inputs, v)
		}
	}

	outputs := make([]gojq.Iter, len(inputs))
	for i, input := range inputs {
		outputs[i] = code.Run(input, values...)
	}

	encoder := json.NewEncoder(os.Stdout)
	for _, output := range outputs {
		for {
			v, ok := output.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			err := encoder.Encode(v)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
	}

}
