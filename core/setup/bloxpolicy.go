package setup

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	"github.com/hashicorp/go-syslog"
	"github.com/miekg/coredns/middleware"
	"github.com/miekg/coredns/middleware/bloxpolicy"
	"github.com/miekg/coredns/server"
	"github.com/miekg/dns"
)

// Log sets up the logging middleware.
func BloxPolicy(c *Controller) (middleware.Middleware, error) {
	rules, endpoint, err := bloxPolicyParse(c)
	if err != nil {
		return nil, err
	}

	// Open the log files for writing when the server starts
	c.Startup = append(c.Startup, func() error {
		for i := 0; i < len(rules); i++ {
			var err error
			var writer io.Writer

			if rules[i].OutputFile == "stdout" {
				writer = os.Stdout
			} else if rules[i].OutputFile == "stderr" {
				writer = os.Stderr
			} else if rules[i].OutputFile == "syslog" {
				writer, err = gsyslog.NewLogger(gsyslog.LOG_INFO, "LOCAL0", "coredns")
				if err != nil {
					return err
				}
			} else {
				var file *os.File
				file, err = os.OpenFile(rules[i].OutputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
				if err != nil {
					return err
				}
				if rules[i].Roller != nil {
					file.Close()
					rules[i].Roller.Filename = rules[i].OutputFile
					writer = rules[i].Roller.GetLogWriter()
				} else {
					writer = file
				}
			}

			rules[i].Log = log.New(writer, "", 0)
		}

		return nil
	})

	return func(next middleware.Handler) middleware.Handler {
		return bloxpolicy.BloxPolicy{Next: next, Rules: rules, ErrorFunc: server.DefaultErrorFunc, Endpoint: endpoint}
	}, nil
}

func bloxPolicyParse(c *Controller) ([]bloxpolicy.Rule, string, error) {
	var rules []bloxpolicy.Rule
	var endpoint string

	for c.Next() {
		args := c.RemainingArgs()

		var logRoller *middleware.LogRoller
		if c.NextBlock() {
			if c.Val() == "rotate" {
				if c.NextArg() {
					if c.Val() == "{" {
						var err error
						logRoller, err = parseRoller(c)
						if err != nil {
							return nil, "", err
						}
						// This part doesn't allow having something after the rotate block
						if c.Next() {
							if c.Val() != "}" {
								return nil, "", c.ArgErr()
							}
						}
					}
				}
			}
		}
		if len(args) == 0 {
			endpoint = bloxpolicy.DefaultEndpoint
			// TODO: get rid of logging code
			// Nothing specified; use defaults
			rules = append(rules, bloxpolicy.Rule{
				NameScope:  ".",
				OutputFile: bloxpolicy.DefaultLogFilename,
				Format:     bloxpolicy.DefaultLogFormat,
				Roller:     logRoller,
			})
		} else if len(args) == 1 {
			fmt.Printf("bloxpolicy has 1 arg: %v\n", args)
			url, error := url.Parse(args[0])
			if error != nil {
				fmt.Printf("[ERROR] bloxpolicy endpoint url is invalid: %v\n", args[0])
				endpoint = bloxpolicy.DefaultEndpoint
			} else {
				endpoint = url.String()
			}

			// TODO: Remove old logging code
			// Hard-code log rules until logging code is removed.
			rules = append(rules, bloxpolicy.Rule{
				NameScope:  ".",
				OutputFile: bloxpolicy.DefaultLogFilename,
				Format:     bloxpolicy.DefaultLogFormat,
				Roller:     logRoller,
			})
		} else {
			// Name scope, output file, and maybe a format specified

			format := bloxpolicy.DefaultLogFormat

			if len(args) > 2 {
				switch args[2] {
				case "{common}":
					format = bloxpolicy.CommonLogFormat
				case "{combined}":
					format = bloxpolicy.CombinedLogFormat
				default:
					format = args[2]
				}
			}

			rules = append(rules, bloxpolicy.Rule{
				NameScope:  dns.Fqdn(args[0]),
				OutputFile: args[1],
				Format:     format,
				Roller:     logRoller,
			})
		}
	}

	return rules, endpoint, nil
}
