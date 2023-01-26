package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"tailscale.com/jsondb"
)

func NewCmdKv() *cobra.Command {
	command := cobra.Command{
		Use:   "kv",
		Short: "Key-value store",
	}

	command.AddCommand(&cobra.Command{
		Use: "get [key]",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}

			data := *db.Data
			value, ok := data[args[0]]
			if !ok {
				return fmt.Errorf("key not found")
			}

			fmt.Println(value)
			return nil
		},
	})

	command.AddCommand(&cobra.Command{
		Use: "set [key] [value]",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}

			data := *db.Data
			data[args[0]] = args[1]

			return db.Save()
		},
	})

	command.AddCommand(&cobra.Command{
		Use: "delete",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}

			data := *db.Data
			delete(data, args[0])

			return db.Save()
		}})

	command.AddCommand(&cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := loadDB()
			if err != nil {
				return err
			}

			data := *db.Data
			for key := range data {
				fmt.Println(key)
			}

			return nil
		},
	})

	return &command
}

func loadDB() (*jsondb.DB[map[string]any], error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	if !isExtensionDir(wd) {
		return nil, fmt.Errorf("not in an extension directory")
	}

	dbPath := path.Join(wd, ".sunbeam", "db.json")

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		err := os.MkdirAll(path.Dir(dbPath), 0755)
		if err != nil {
			return nil, err
		}
		os.WriteFile(dbPath, []byte("{}"), 0644)
		if err != nil {
			return nil, err
		}
	}

	return jsondb.Open[map[string]any](dbPath)
}

func isExtensionDir(dir string) bool {
	_, err := os.Stat(path.Join(dir, "sunbeam.yml"))

	return !os.IsNotExist(err)
}
