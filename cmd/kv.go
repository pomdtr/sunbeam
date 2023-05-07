package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

const (
	kvFile = "kv.json"
)

func NewKvCmd(extensionRoot string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "kv",
		Short:   "Key/Value operations",
		GroupID: coreGroupID,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			kvPath, _ := cmd.Flags().GetString("db")
			if kvPath != "" {
				return nil
			}

			return fmt.Errorf("db flag is required if not set in environment")
		},
	}

	cmd.AddCommand(NewKvGetCmd(extensionRoot))
	cmd.AddCommand(NewKvSetCmd(extensionRoot))
	cmd.AddCommand(NewKvListCmd(extensionRoot))
	cmd.AddCommand(NewKvDeleteCmd(extensionRoot))

	cmd.PersistentFlags().String("db", os.Getenv("SUNBEAM_KV_PATH"), "Path to the key/value store")

	return cmd
}

func NewKvGetCmd(extensionRoot string) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get a value from the key/value store",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kvPath, _ := cmd.Flags().GetString("db")

			db, err := OpenDB(kvPath)
			if err != nil {
				return err
			}

			val, ok := db.Data[args[0]]
			if !ok {
				return fmt.Errorf("key %s does not exist", args[0])
			}

			fmt.Println(val)
			return nil
		},
	}
}

func NewKvSetCmd(extensionRoot string) *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Set a value in the key/value store",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kvPath, _ := cmd.Flags().GetString("db")

			db, err := OpenDB(kvPath)
			if err != nil {
				return err
			}

			key := args[0]
			var val string
			if len(args) == 2 {
				val = args[1]
			} else {
				if isatty.IsTerminal(os.Stdin.Fd()) {
					return fmt.Errorf("value is required when not piping")
				}

				bs, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				val = string(bs)
			}

			db.Data[key] = val

			if err := db.Save(); err != nil {
				return err
			}

			return nil
		},
	}
}

func NewKvListCmd(extensionRoot string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all key/value pairs",
		RunE: func(cmd *cobra.Command, args []string) error {
			kvPath, _ := cmd.Flags().GetString("db")

			db, err := OpenDB(kvPath)
			if err != nil {
				return err
			}

			delimiter, _ := cmd.Flags().GetString("delimiter")
			keysOnly, _ := cmd.Flags().GetBool("keys-only")
			valuesOnly, _ := cmd.Flags().GetBool("values-only")
			for k, v := range db.Data {
				if keysOnly {
					fmt.Println(k)
				} else if valuesOnly {
					fmt.Println(v)
				} else {
					fmt.Printf("%s%s%s\n", k, delimiter, v)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringP("delimiter", "d", "\t", "Delimiter to separate keys and values")
	cmd.Flags().BoolP("keys-only", "k", false, "Only print keys")
	cmd.Flags().BoolP("values-only", "v", false, "Only print values")

	return cmd
}

func NewKvDeleteCmd(extensionRoot string) *cobra.Command {
	return &cobra.Command{
		Use:   "delete",
		Short: "Delete a key/value pair",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			kvPath, _ := cmd.Flags().GetString("db")

			db, err := OpenDB(kvPath)
			if err != nil {
				return err
			}

			if _, ok := db.Data[args[0]]; !ok {
				return fmt.Errorf("key %s does not exist", args[0])
			}

			delete(db.Data, args[0])

			if err := db.Save(); err != nil {
				return err
			}

			return nil
		},
	}
}

type DB struct {
	Data map[string]string
	path string
}

// Open opens the database at path, creating it with a zero value if necessary.
func OpenDB(path string) (*DB, error) {
	bs, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return &DB{
			Data: make(map[string]string),
			path: path,
		}, nil
	} else if err != nil {
		return nil, err
	}

	var val map[string]string
	if err := json.Unmarshal(bs, &val); err != nil {
		return nil, err
	}

	return &DB{
		Data: val,
		path: path,
	}, nil
}

// Save writes db.Data back to disk.
func (db DB) Save() error {
	kvDir := filepath.Dir(db.path)
	if err := os.MkdirAll(kvDir, 0700); err != nil {
		return err
	}

	bs, err := json.MarshalIndent(db.Data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(db.path, bs, 0600)
}
