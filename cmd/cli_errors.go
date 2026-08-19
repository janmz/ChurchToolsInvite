package cmd

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

type usageError struct {
	cmd *cobra.Command
	err error
}

func (e *usageError) Error() string {
	return e.err.Error()
}

func (e *usageError) Unwrap() error {
	return e.err
}

func usageErrorf(cmd *cobra.Command, format string, args ...any) error {
	return &usageError{cmd: cmd, err: fmt.Errorf(format, args...)}
}

var (
	reUnknownFlag    = regexp.MustCompile(`^unknown flag: (.+)$`)
	reUnknownCommand = regexp.MustCompile(`^unknown command "(.+)" for "(.+)"$`)
	reRequiredFlag   = regexp.MustCompile(`^required flag\(s\) "(.+)" not set$`)
	reInvalidArg     = regexp.MustCompile(`^invalid argument "(.+)" for "(.+)" flag: (.+)$`)
)

var flagDisplayNames = map[string]string{
	"csv": "--csv (-f)",
}

func translateCLIError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	if m := reUnknownFlag.FindStringSubmatch(msg); len(m) == 2 {
		return fmt.Errorf("unbekannte Option %s", m[1])
	}
	if m := reUnknownCommand.FindStringSubmatch(msg); len(m) == 3 {
		return fmt.Errorf("unbekannter Befehl %q (Programm %q)", m[1], m[2])
	}
	if m := reRequiredFlag.FindStringSubmatch(msg); len(m) == 2 {
		if label, ok := flagDisplayNames[m[1]]; ok {
			return fmt.Errorf("Pflichtoption %s fehlt", label)
		}
		return fmt.Errorf("Pflichtoption --%s fehlt", m[1])
	}
	if m := reInvalidArg.FindStringSubmatch(msg); len(m) == 4 {
		detail := simplifyParseDetail(m[3])
		return fmt.Errorf("ungültiger Wert %q für Option %s: %s", m[1], m[2], detail)
	}

	return err
}

func simplifyParseDetail(detail string) string {
	if idx := strings.Index(detail, ": "); idx >= 0 {
		detail = detail[idx+2:]
	}
	switch {
	case strings.Contains(detail, "invalid syntax"):
		return "keine Zahl"
	case strings.Contains(detail, "ParseBool"):
		return "kein Wahrheitswert (true/false)"
	default:
		return detail
	}
}

func flagErrorFunc(cmd *cobra.Command, err error) error {
	applyGermanFlagLabels(cmd)
	_ = cmd.Usage()
	return translateCLIError(err)
}

func configureGermanCLI(cmd *cobra.Command) {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SetFlagErrorFunc(flagErrorFunc)
	for _, sub := range cmd.Commands() {
		configureGermanCLI(sub)
	}
}

func finalizeCLIError(err error) error {
	if err == nil {
		return nil
	}

	var uErr *usageError
	if errors.As(err, &uErr) {
		applyGermanFlagLabels(uErr.cmd)
		_ = uErr.cmd.Usage()
		return uErr.err
	}

	translated := translateCLIError(err)
	msg := err.Error()

	if reUnknownCommand.MatchString(msg) {
		applyGermanFlagLabels(rootCmd)
		_ = rootCmd.Usage()
	}

	return translated
}
