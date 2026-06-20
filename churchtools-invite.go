package main

/*
 * churchtools-invite: A tool to invite persons to ChurchTools, based on imported data.
 *
 * This command-line application is used to automate inviting people into a ChurchTools instance.
 * It processes person data exported from other systems or spreadsheets, checks for existing user records,
 * optionally synchronizes email addresses, and handles permissions and status updates required for ChurchTools invitations.
 * The tool supports both main and sub-instances (multi-tenancy) via OAuth, and can be used in dry-run mode for verification
 * before actually sending invitations. Its main purpose is to streamline the onboarding process for churches and organizations
 * using ChurchTools for managing their contacts and communications.
 *
 *
 * Version: 2.6.2.49 (in version.go zu ändern)
 *
 * Author: Jan Neuhaus, VAYA Consulting, https://vaya-consulting.de/development
 *
 * Repository: https://github.com/janmz/ChurchToolsInvite
 *
 * ChangeLog:
 *  20.06.26	2.6.2	Fixes: auto-signup of groups will be used, groups displayed with whoami, pre_join_groups corrected, no invitation to registered users, --interactive has new filter for invitation status
 *  20.06.26	2.6.0	Feature: Now only uninvited persons are exported unless --invited is given. --campus-id becomes --campus and accepts (partial) names, csv includes campus and status
 *  17.06.26	2.4.0	Fixed: Replacement of e-mail-addresses now working and tested
 *  17.06.26	2.3.2	Fixed: Windows build including icon - worked!
 *  17.06.26	2.3.0	Feature: Updated documentation, fixed small vulnerabilities, expaned test code
 *  17.06.26	2.2.0	Full support of main and sub instances using OAuth
 *  17.06.26	2.1.0	Including executables for Windows, Linux and macOS
 *  17.06.26	2.0.0	Published on GitHub
 *  16.06.26	1.0.6	First working version
 *  15.06.26	1.0.0	Initial version
 *
 * (c)2026 Jan Neuhaus, VAYA Consulting
 *
 */

import (
	"fmt"
	"os"

	cmd "github.com/janmz/churchtools-invite/cmd"
)

func main() {
	if err := cmd.Execute(fmt.Sprintf("%s (%s)", Version, BuildTime)); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
