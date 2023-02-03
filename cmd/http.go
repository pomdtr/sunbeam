package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-isatty"
	httpie "github.com/nojima/httpie-go"
	"github.com/nojima/httpie-go/flags"
	"github.com/nojima/httpie-go/input"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type httpOptions struct {
	flags.OptionSet
	Print       string
	Verbose     bool
	Body        bool
	Headers     bool
	IgnoreStdin bool
	OutputFile  string
	Verify      string
	TimeOut     string
	Auth        string
	Pretty      string
	Version     bool
	Licence     bool
}

func NewCmdHttp() *cobra.Command {
	var optionSet httpOptions
	cmd := cobra.Command{
		Use:   "http [METHOD] URL [REQUEST_ITEM ...]",
		Short: "User-friendly curl replacement inspired by HTTPie",
		PreRunE: func(cmd *cobra.Command, args []string) (err error) {

			if !optionSet.IgnoreStdin && !isatty.IsTerminal(os.Stdin.Fd()) {
				optionSet.InputOptions.ReadStdin = true
			}

			err = parsePrintFlag(&optionSet)
			if err != nil {
				return
			}

			optionSet.ExchangeOptions.Timeout, err = parseDurationOrSeconds(optionSet.TimeOut)
			if err != nil {
				return
			}

			if err := parsePretty(&optionSet); err != nil {
				return err
			}

			verifyFlag := strings.ToLower(optionSet.Verify)
			switch verifyFlag {
			case "no":
				optionSet.ExchangeOptions.SkipVerify = true
			case "yes":
			case "":
				optionSet.ExchangeOptions.SkipVerify = false
			default:
				return fmt.Errorf("%s", "Verify flag must be 'yes' or 'no'")
			}

			// Parse --auth
			if optionSet.Auth != "" {
				username, password := parseAuth(optionSet.Auth)

				if password == nil {
					return fmt.Errorf("%s", "Password is required")
				}

				optionSet.ExchangeOptions.Auth.Enabled = true
				optionSet.ExchangeOptions.Auth.UserName = username
				optionSet.ExchangeOptions.Auth.Password = *password
			}

			return
		},
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			inputOptions := optionSet.InputOptions
			exchangeOptions := optionSet.ExchangeOptions
			outputOptions := optionSet.OutputOptions

			// Parse positional arguments
			in, err := input.ParseArgs(args, os.Stdin, &inputOptions)
			if _, ok := errors.Cause(err).(*input.UsageError); ok {
				return err
			}
			if err != nil {
				return err
			}

			// Send request and receive response
			status, err := httpie.Exchange(in, &exchangeOptions, &outputOptions)
			if err != nil {
				return err
			}

			if exchangeOptions.CheckStatus && 300 <= status && status < 600 {
				return fmt.Errorf("status: %d", status)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&optionSet.InputOptions.JSON, "json", "j", false, "data items are serialized as JSON (default)")
	cmd.Flags().BoolVarP(&optionSet.InputOptions.Form, "form", "f", false, "data items are serialized as form fields")
	cmd.Flags().StringVarP(&optionSet.Print, "print", "p", "", "specifies what the output should contain (HBhb)")
	cmd.Flags().BoolVarP(&optionSet.Verbose, "verbose", "v", false, "print the request as well as the response. shortcut for --print=HBhb")
	cmd.Flags().BoolVarP(&optionSet.Headers, "headers", "H", false, "print only the request headers. shortcut for --print=h")
	cmd.Flags().BoolVarP(&optionSet.Body, "body", "b", false, "print only response body. shourtcut for --print=b")
	cmd.Flags().BoolVar(&optionSet.InputOptions.ReadStdin, "ignore-stdin", false, "do not attempt to read stdin")
	cmd.Flags().BoolVarP(&optionSet.OutputOptions.Download, "download", "d", false, "download file")
	cmd.Flags().BoolVar(&optionSet.OutputOptions.Overwrite, "overwrite", false, "overwrite existing file")
	cmd.Flags().BoolVar(&optionSet.ExchangeOptions.ForceHTTP1, "http1", false, "force HTTP/1.1 protocol")
	cmd.Flags().StringVarP(&optionSet.OutputFile, "output", "o", "", "output file")
	cmd.Flags().StringVar(&optionSet.Verify, "verify", "", "verify Host SSL certificate, 'yes' or 'no' ('yes' by default, uppercase is also working)")
	cmd.Flags().StringVar(&optionSet.TimeOut, "timeout", "30s", "timeout seconds that you allow the whole operation to take")
	cmd.Flags().BoolVar(&optionSet.ExchangeOptions.CheckStatus, "check-status", false, "Also check the HTTP status code and exit with an error if the status indicates one")
	cmd.Flags().StringVarP(&optionSet.Auth, "auth", "a", "", "colon-separated username and password for authentication")
	cmd.Flags().StringVar(&optionSet.Pretty, "pretty", "", "controls output formatting (all, format, none)")
	cmd.Flags().BoolVarP(&optionSet.ExchangeOptions.FollowRedirects, "follow", "F", false, "follow 30x Location redirects")

	return &cmd

}

func parsePrintFlag(options *httpOptions) error {
	if options.Print == "" {
		if options.Headers {
			options.OutputOptions.PrintResponseHeader = true
		} else if options.Body {
			options.OutputOptions.PrintResponseBody = true
		} else if options.Verbose {
			options.OutputOptions.PrintRequestBody = true
			options.OutputOptions.PrintRequestHeader = true
			options.OutputOptions.PrintResponseHeader = true
			options.OutputOptions.PrintResponseBody = true
		} else if isatty.IsTerminal(os.Stdout.Fd()) {
			options.OutputOptions.PrintResponseHeader = true
			options.OutputOptions.PrintResponseBody = true
		} else {
			options.OutputOptions.PrintResponseBody = true
		}
	} else { // --print is specified
		for _, c := range options.Print {
			switch c {
			case 'H':
				options.OutputOptions.PrintRequestHeader = true
			case 'B':
				options.OutputOptions.PrintRequestBody = true
			case 'h':
				options.OutputOptions.PrintResponseHeader = true
			case 'b':
				options.OutputOptions.PrintResponseBody = true
			default:
				return errors.Errorf("invalid char in --print value (must be consist of HBhb): %c", c)
			}
		}
	}
	return nil
}

func parseDurationOrSeconds(timeout string) (time.Duration, error) {
	if _, err := strconv.Atoi(timeout); err == nil {
		timeout += "s"
	}
	d, err := time.ParseDuration(timeout)
	if err != nil {
		return time.Duration(0), errors.Errorf("Value of --timeout must be a number or duration string: %v", timeout)
	}
	return d, nil
}

func parsePretty(options *httpOptions) error {
	switch options.Pretty {
	case "":
		stdoutIsTerminal := isatty.IsTerminal(os.Stdout.Fd())
		options.OutputOptions.EnableFormat = stdoutIsTerminal
		options.OutputOptions.EnableColor = stdoutIsTerminal
	case "all":
		options.OutputOptions.EnableFormat = true
		options.OutputOptions.EnableColor = true
	case "none":
		options.OutputOptions.EnableFormat = false
		options.OutputOptions.EnableColor = false
	case "format":
		options.OutputOptions.EnableFormat = true
		options.OutputOptions.EnableColor = false
	case "colors":
		return errors.New("--pretty=colors is not implemented")
	default:
		return errors.Errorf("unknown value of --pretty: %s", options.Pretty)
	}
	return nil
}

func parseAuth(authFlag string) (string, *string) {
	colonIndex := strings.Index(authFlag, ":")
	if colonIndex == -1 {
		return authFlag, nil
	}
	password := authFlag[colonIndex+1:]
	return authFlag[:colonIndex], &password
}
