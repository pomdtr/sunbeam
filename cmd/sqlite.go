package cmd

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/spf13/cobra"
)

func NewCmdSQL() *cobra.Command {
	return &cobra.Command{
		Use:     "sql <database> <query>",
		Short:   "Run a SQL query on a SQLite database",
		GroupID: "core",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			db, err := sql.Open("sqlite", args[0])
			if err != nil {
				return err
			}

			rows, err := db.Query(args[1])
			if err != nil {
				return err
			}

			for rows.Next() {
				var i any
				if err := rows.Scan(&i); err != nil {
					return err
				}
				fmt.Println(i)
			}

			if err := rows.Err(); err != nil {
				return err
			}

			if err := db.Close(); err != nil {
				return err
			}

			return nil
		},
	}

}
