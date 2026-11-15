/*
 * settings.go - Copyright (c) 2023-2026 - Olivier Poncet
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 2 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package app

import (
    "errors"
    "flag"
    "io"
    "log/slog"
    "path/filepath"
    "strings"

    "timesheet/internal/activity"
)

// ---------------------------------------------------------------------------
// Settings declaration
// ---------------------------------------------------------------------------

type Settings struct {
    command        string
    arguments      []string
    database       string
    template       string
    config         string
    logLevel       slog.Level
    listen         string
    origins        *OptionList
    trustedProxies *OptionList
    tlsCert        string
    tlsKey         string
    profile        string
    provisional    bool
    from           string
    to             string
    dryRun         bool
    replace        bool
    validate       bool
    explicit       map[string]bool
}

var ErrUsage = errors.New("invalid usage")

// ---------------------------------------------------------------------------
// Settings implementation
// ---------------------------------------------------------------------------

func NewSettings(args []string) (*Settings, error) {
    s := &Settings{origins: NewOptionList(), trustedProxies: NewOptionList(), explicit: map[string]bool{}}
    positional, options := []string{}, []string{}
    for _, arg := range args {
        if strings.HasPrefix(arg, "-") {
            options = append(options, arg)
        } else {
            positional = append(positional, arg)
        }
    }
    flags := flag.NewFlagSet("timesheet", flag.ContinueOnError)
    flags.SetOutput(io.Discard)
    level := flags.String("log-level", "info", "")
    flags.StringVar(&s.database, "database", "", "")
    flags.StringVar(&s.template, "template", "", "")
    flags.StringVar(&s.config, "config", "", "")
    flags.StringVar(&s.listen, "listen", "", "")
    flags.Var(s.origins, "origin", "")
    flags.Var(s.trustedProxies, "trusted-proxy", "")
    flags.StringVar(&s.tlsCert, "tls-cert", "", "")
    flags.StringVar(&s.tlsKey, "tls-key", "", "")
    flags.StringVar(&s.profile, "profile", "USER", "")
    flags.BoolVar(&s.provisional, "provisional", false, "")
    flags.StringVar(&s.from, "from", "", "")
    flags.StringVar(&s.to, "to", "", "")
    flags.BoolVar(&s.dryRun, "dry-run", false, "")
    flags.BoolVar(&s.replace, "replace", false, "")
    flags.BoolVar(&s.validate, "validate", false, "")
    if err := flags.Parse(options); err != nil {
        if errors.Is(err, flag.ErrHelp) {
            return nil, flag.ErrHelp
        }
        return nil, NewUsageError(err.Error())
    }
    if flags.NArg() > 0 {
        return nil, NewUsageError("unexpected argument " + flags.Arg(0))
    }
    flags.Visit(s.mark)
    if err := s.logLevel.UnmarshalText([]byte(*level)); err != nil {
        return nil, NewUsageError("invalid --log-level")
    }
    if len(positional) == 0 {
        return nil, flag.ErrHelp
    }
    s.command = positional[0]
    s.arguments = positional[1:]
    if (s.command == "config") || (s.command == "user") || (s.command == "session") {
        if len(s.arguments) == 0 {
            return nil, NewUsageError(s.command + ": missing subcommand")
        }
        s.command += " " + s.arguments[0]
        s.arguments = s.arguments[1:]
    }
    return s, s.check()
}

func (s *Settings) mark(f *flag.Flag) {
    s.explicit[f.Name] = true
}

func (s *Settings) check() error {
    counts := map[string]int{"help": 0, "version": 0, "init": 0, "run": 0, "config show": 0, "config set": 2, "user add": 1, "user profile": 2, "user passwd": 1, "user del": 1, "user list": 0, "session clear": 0, "import": 2}
    count, known := counts[s.command]
    if !known {
        return NewUsageError("unknown command " + s.command)
    }
    switch s.command {
        case "config show":
            if len(s.arguments) > 1 {
                return NewUsageError(s.command + ": wrong number of arguments")
            }
        case "config set":
            if (len(s.arguments) < count) || ((len(s.arguments) % 2) != 0) {
                return NewUsageError(s.command + ": pairs KEY VALUE expected")
            }
        default:
            if len(s.arguments) != count {
                return NewUsageError(s.command + ": wrong number of arguments")
            }
    }
    allowed := map[string][]string{"init": {"template"}, "run": {"listen", "origin", "trusted-proxy", "tls-cert", "tls-key"}, "user add": {"profile", "provisional"}, "import": {"from", "to", "dry-run", "replace", "validate"}}
    for name := range s.explicit {
        if (name == "database") || (name == "config") || (name == "log-level") {
            continue
        }
        permitted := false
        for _, option := range allowed[s.command] {
            permitted = permitted || (option == name)
        }
        if !permitted {
            return NewUsageError("--" + name + " does not apply to " + s.command)
        }
    }
    if s.explicit["database"] && !activity.ValidDatabasePath(s.database) {
        return NewUsageError("--database names a file, not a directory, . or .., whose name does not end with " + activity.LockSuffix + ", " + activity.SessionsSuffix + " or " + activity.TemporarySuffix + " in any case")
    }
    if strings.HasPrefix(s.command, "config ") && !s.explicit["config"] {
        return NewUsageError(s.command + ": --config is required")
    }
    if s.explicit["config"] && ((s.config == "") || strings.HasSuffix(s.config, "/")) {
        return NewUsageError("--config names a file")
    }
    if (s.command == "init") && ((s.template == "") || strings.HasSuffix(s.template, "/")) {
        return NewUsageError("init: --template names the configuration template, a file")
    }
    if (s.tlsCert == "") != (s.tlsKey == "") {
        return NewUsageError("--tls-cert and --tls-key go together")
    }
    if _, err := activity.NewTrustedProxies(s.trustedProxies.Values()); err != nil {
        return NewUsageError("invalid --trusted-proxy")
    }
    return nil
}

func (s *Settings) Command() string {
    return s.command
}

func (s *Settings) Argument(i int) string {
    return s.arguments[i]
}

func (s *Settings) Arguments() []string {
    return s.arguments
}

func (s *Settings) Database() string {
    return s.database
}

func (s *Settings) ResolveDatabase(configured string) error {
    if !s.explicit["database"] {
        s.database = configured
    }
    if s.database == "" {
        return NewUsageError(s.command + ": no database: give --database or set database.path in --config")
    }
    absolute, err := filepath.Abs(s.database)
    if err != nil {
        return err
    }
    s.database = absolute
    return nil
}

func (s *Settings) ExplicitLogLevel() bool {
    return s.explicit["log-level"]
}

func (s *Settings) Template() string {
    return s.template
}

func (s *Settings) LogLevel() slog.Level {
    return s.logLevel
}

func (s *Settings) Listen() string {
    return s.listen
}

func (s *Settings) Origins() []string {
    return s.origins.Values()
}

func (s *Settings) Config() string {
    return s.config
}

func (s *Settings) HasConfig() bool {
    return s.explicit["config"]
}

func (s *Settings) Paths() string {
    paths := s.quoted("--database=" + s.database)
    if s.HasConfig() {
        config, err := filepath.Abs(s.config)
        if err != nil {
            config = s.config
        }
        paths += " " + s.quoted("--config=" + config)
    }
    return paths
}

func (s *Settings) quoted(argument string) string {
    for _, letter := range argument {
        if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-./=:@%+,", letter) {
            return "'" + strings.ReplaceAll(argument, "'", "'\\''") + "'"
        }
    }
    return argument
}

func (s *Settings) TrustedProxies() []string {
    return s.trustedProxies.Values()
}

func (s *Settings) TLS() bool {
    return s.tlsCert != ""
}

func (s *Settings) TLSPair() (string, string) {
    return s.tlsCert, s.tlsKey
}

func (s *Settings) Profile() string {
    return s.profile
}

func (s *Settings) Provisional() bool {
    return s.provisional
}

func (s *Settings) Period() (string, string) {
    return s.from, s.to
}

func (s *Settings) DryRun() bool {
    return s.dryRun
}

func (s *Settings) Replace() bool {
    return s.replace
}

func (s *Settings) Validate() bool {
    return s.validate
}

// ---------------------------------------------------------------------------
// End-Of-File
// ---------------------------------------------------------------------------
