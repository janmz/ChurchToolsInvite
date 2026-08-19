package cmd

const usageTemplateDE = `Verwendung:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [Befehl]{{end}}{{if gt (len .Aliases) 0}}

Aliase:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Beispiel:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Verfügbare Befehle:{{$parent := .}}{{range .Commands}}{{if (and (eq .GroupID "") (or .IsAvailableCommand (eq .Name "help")))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Optionen:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Globale Optionen:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Zusätzliche Hilfe-Befehle:{{range .Commands}}{{if eq .GroupID "help"}}{{if .IsAvailableCommand}}
  {{rpad .CommandPath .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Weitere Hilfe: {{.CommandPath}} [Befehl] --help
{{end}}`

const helpTemplateDE = `{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}`
