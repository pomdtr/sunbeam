package cmd

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func NewCmdToken() *cobra.Command {
	flags := struct {
		size int
	}{}

	cmd := &cobra.Command{
		Use:     "token",
		Short:   "Generate a bearer token",
		GroupID: coreGroupID,
		RunE: func(cmd *cobra.Command, args []string) error {
			randomBytes := make([]byte, flags.size)

			_, err := rand.Read(randomBytes)
			if err != nil {
				return err
			}

			for i, b := range randomBytes {
				randomBytes[i] = chars[b%byte(len(chars))]
			}

			token := string(randomBytes)

			if isatty.IsTerminal(os.Stdout.Fd()) {
				fmt.Println(token)
			} else {
				fmt.Print(token)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&flags.size, "size", 32, "Size of the token")

	return cmd
}
