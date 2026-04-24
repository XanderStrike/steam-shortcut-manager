/*
MIT License

Copyright © 2026 Alex Standke <xanderstrike at gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"time"

	"github.com/shadowblip/steam-shortcut-manager/pkg/steam"
	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup Steam shortcuts to the current directory",
	Long:  `Backs up the Steam shortcuts VDF file to the current directory with a date-stamped filename (e.g. shortcuts.2026-01-01.vdf)`,
	Run: func(cmd *cobra.Command, args []string) {
		format := rootCmd.PersistentFlags().Lookup("output").Value.String()

		// Determine which user to back up
		userID, _ := cmd.Flags().GetString("user")
		users, err := steam.GetUsers()
		if err != nil {
			ExitError(err, format)
		}

		// If no user specified, back up all users that have shortcuts
		if userID == "" {
			for _, u := range users {
				if !steam.HasShortcuts(u) {
					continue
				}
				if err := backupUserShortcuts(u, format); err != nil {
					ExitError(err, format)
				}
			}
			return
		}

		// Validate the specified user exists
		if !contains(users, userID) {
			ExitError(fmt.Errorf("user %s not found", userID), format)
		}
		if !steam.HasShortcuts(userID) {
			ExitError(fmt.Errorf("user %s has no shortcuts file", userID), format)
		}
		if err := backupUserShortcuts(userID, format); err != nil {
			ExitError(err, format)
		}
	},
}

func backupUserShortcuts(user string, format string) error {
	shortcutsPath, err := steam.GetShortcutsPath(user)
	if err != nil {
		return err
	}

	// Open the source file
	src, err := os.Open(shortcutsPath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", shortcutsPath, err)
	}
	defer src.Close()

	// Build the destination filename: shortcuts.YYYY-MM-DD.vdf
	dateStr := time.Now().Format("2006-01-02")
	destName := fmt.Sprintf("shortcuts.%s.vdf", dateStr)

	// If multiple users, prefix with user ID to avoid collisions
	users, _ := steam.GetUsers()
	if len(users) > 1 {
		destName = fmt.Sprintf("shortcuts.%s.%s.vdf", user, dateStr)
	}

	destPath := path.Join(".", destName)

	// Create the destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", destPath, err)
	}
	defer dst.Close()

	// Copy the file contents
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("failed to copy shortcuts: %w", err)
	}

	// Print the output
	switch format {
	case "term":
		fmt.Printf("Backed up %s -> %s\n", shortcutsPath, destPath)
	case "json":
		out, _ := json.MarshalIndent(map[string]string{
			"user":   user,
			"source": shortcutsPath,
			"dest":   destPath,
		}, "", "  ")
		fmt.Println(string(out))
	default:
		panic("unknown output format: " + format)
	}

	return nil
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringP("user", "u", "", "Steam user ID to back up (backs up all users if not specified)")
}
