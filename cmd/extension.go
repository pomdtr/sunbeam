package cmd

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/cli/browser"
	"github.com/pomdtr/sunbeam/internal"
	"github.com/spf13/cobra"
)

func NewExtensionCmd(extensionMap internal.Extensions) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "extension",
		Short:   "Manage extensions",
		GroupID: coreGroupID,
	}

	cmd.AddCommand(NewExtensionAddCmd(extensionMap))
	cmd.AddCommand(NewExtensionOpenCmd(extensionMap))
	cmd.AddCommand(NewExtensionCmdList(extensionMap))
	cmd.AddCommand(NewExtensionUpgrade(extensionMap))
	cmd.AddCommand(NewExtensionRenameCmd(extensionMap))
	cmd.AddCommand(NewExtensionCmdRemove(extensionMap))

	return cmd
}

func parseOrigin(origin string) (*url.URL, error) {
	url, err := url.Parse(origin)
	if err != nil {
		return nil, err
	}

	if url.Scheme == "" {
		url.Scheme = "file"
	}

	if url.Scheme != "file" && url.Scheme != "http" && url.Scheme != "https" {
		return nil, fmt.Errorf("invalid origin: %s", origin)
	}

	if url.Scheme == "file" && !filepath.IsAbs(url.Path) {
		abs, err := filepath.Abs(url.Path)
		if err != nil {
			return nil, err
		}

		url.Path = abs
	}

	return url, nil
}

func NewExtensionAddCmd(extensions internal.Extensions) *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <origin>",
		Short: "Add an extension",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := extensions.Get(args[0]); err == nil {
				return fmt.Errorf("extension %s already exists", args[0])
			}

			return nil
		},
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			originUrl, err := parseOrigin(args[1])
			if err != nil {
				return err
			}

			cmd.PrintErrf("Loading manifest from %s\n", originUrl.String())
			manifest, err := internal.LoadManifest(originUrl)
			if err != nil {
				return err
			}

			extensions.Add(args[0], internal.Extension{
				Origin:   originUrl.String(),
				Manifest: manifest,
			})

			if err := extensions.Save(); err != nil {
				return err
			}

			cmd.PrintErrln("Extension added successfully!")
			return nil
		},
	}
}

func NewExtensionOpenCmd(extensions internal.Extensions) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "open <name>",
		Short:     "Open an extension's homepage",
		Args:      cobra.ExactArgs(1),
		ValidArgs: extensions.List(),
		RunE: func(cmd *cobra.Command, args []string) error {
			extension, err := extensions.Get(args[0])
			if err != nil {
				return err
			}

			if extension.Homepage == "" {
				cmd.PrintErrf("Extension %s does not have a homepage\n", args[0])
				return nil
			}

			if err := browser.OpenURL(extension.Homepage); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func NewExtensionCmdList(extensions internal.Extensions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List extensions",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			for name, extension := range extensions.Map() {
				fmt.Printf("%s\t%s\t%s\n", name, extension.Title, extension.Origin)
			}

			return nil
		},
	}

	return cmd
}

func NewExtensionCmdRemove(extensions internal.Extensions) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "remove <name>",
		Short:     "Remove an extension",
		ValidArgs: extensions.List(),
		Args:      cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if _, err := extensions.Get(args[0]); err != nil {
				return fmt.Errorf("extension %s does not exist", args[0])
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			if err := extensions.Remove(name); err != nil {
				return err
			}

			if err := extensions.Save(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}

func NewExtensionUpgrade(extensions internal.Extensions) *cobra.Command {
	flags := struct {
		all bool
	}{}

	cmd := &cobra.Command{
		Use:       "upgrade <name>",
		Short:     "Update an extension",
		Args:      cobra.ArbitraryArgs,
		ValidArgs: extensions.List(),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if flags.all && len(args) > 0 {
				return fmt.Errorf("cannot use --all and specify extensions")
			}

			if !flags.all && len(args) == 0 {
				return fmt.Errorf("must specify an extension or use --all")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			toUpgrade := args
			if flags.all {
				toUpgrade = extensions.List()
			}

			for _, name := range toUpgrade {
				extension, err := extensions.Get(name)
				if err != nil {
					return err
				}

				origin, err := parseOrigin(extension.Origin)
				if err != nil {
					return err
				}

				cmd.PrintErrf("Extracting manifest from %s\n", origin.String())
				manifest, err := internal.LoadManifest(origin)
				if err != nil {
					return err
				}

				extension.Manifest = manifest
				if err := extensions.Update(name, extension); err != nil {
					return err
				}

				if err := extensions.Save(); err != nil {
					return err
				}

				cmd.PrintErrf("Extension %s upgraded successfully!\n", name)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&flags.all, "all", false, "Upgrade all extensions")

	return cmd
}

func NewExtensionRenameCmd(extensions internal.Extensions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rename <old-name> <new-name>",
		Short: "Rename an extension",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName := args[0]
			newName := args[1]

			if err := extensions.Rename(oldName, newName); err != nil {
				return err
			}

			if err := extensions.Remove(oldName); err != nil {
				return err
			}

			if err := extensions.Save(); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
