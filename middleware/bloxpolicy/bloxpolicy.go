// Package bloxpolicy implements policy-based firewalling
package bloxpolicy

import (
	"fmt"
	"log"
//	"time"

	"github.com/miekg/coredns/middleware"
//	"github.com/miekg/coredns/middleware/metrics"

	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// BloxPolicy is a policy-based firewall
type BloxPolicy struct {
	Endpoint	string
	Next		middleware.Handler
	Rules		[]Rule
	ErrorFunc func(dns.ResponseWriter, *dns.Msg, int) // failover error handler
}


func (p BloxPolicy) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := middleware.State{W: w, Req: r}

	clientIP := state.IP()
	fmt.Printf("[DEBUG] bloxpolicy enter...   client IP: %v\n", clientIP)
	port, err := state.Port()
	fmt.Printf("[DEBUG] bloxpolicy enter... client port: %v\n", port)

	fmt.Printf("bloxpolicy enter... endpoint: %v\n", p.Endpoint)
	fmt.Printf("bloxpolicy enter... state: %v\n", state)
	fmt.Printf("bloxpolicy enter...     w: %v\n", w)
	fmt.Printf("bloxpolicy enter...     r: %v\n", r)

	clientAllowed, err := p.IsClientAllowed(clientIP)
	if ! clientAllowed {
		fmt.Printf("[INFO] client lookup blocked by policy engine for client %v\n", clientIP)
		return dns.RcodeRefused, nil
	}

	/*
	for _, rule := range p.Rules {
		if middleware.Name(rule.NameScope).Matches(state.Name()) {
			responseRecorder := middleware.NewResponseRecorder(w)
			rcode, err := p.Next.ServeDNS(ctx, responseRecorder, r)

			fmt.Printf("bloxpolicy outbound rcode: %v\n", rcode)
			fmt.Printf("bloxpolicy outbound responseRecorder: %v\n", responseRecorder)


			if rcode > 0 {
				// There was an error up the chain, but no response has been written yet.
				// The error must be handled here so the log entry will record the response size.
				if p.ErrorFunc != nil {
					p.ErrorFunc(responseRecorder, r, rcode)
				} else {
					rc := middleware.RcodeToString(rcode)

					answer := new(dns.Msg)
					answer.SetRcode(r, rcode)
					state.SizeAndDo(answer)

					metrics.Report(metrics.Dropped, state.Proto(), rc, answer.Len(), time.Now())
					w.WriteMsg(answer)

					fmt.Printf("answer: %v\n", answer)
				}
				rcode = 0
			}
			rep := middleware.NewReplacer(r, responseRecorder, CommonLogEmptyValue)
			rule.Log.Println(rep.Replace(rule.Format))

			fmt.Printf("here2: rcode %v\n", rcode)
			fmt.Printf("here2: w %v\n", w)
			fmt.Printf("here2: w.RemoteAddr %v\n", w.RemoteAddr())
			fmt.Printf("here2: r %v\n", r)
			fmt.Printf("here2: r.Answer %v\n", r.Answer)
			fmt.Printf("here2: r.Extra %v\n", r.Extra)

			fmt.Printf("#######################   ####\n")
			return rcode, err

		}
	}
*/

	rcode, err := p.Next.ServeDNS(ctx, w, r)

	fmt.Printf("here: rcode %v\n", rcode)
	fmt.Printf("here: w %v\n", w)
	fmt.Printf("here: r %v\n", r)

	fmt.Printf("####    #############   ####\n")
	return rcode, err
}


// Rule configures the logging middleware.
type Rule struct {
	NameScope  string
	OutputFile string
	Format     string
	Log        *log.Logger
	Roller     *middleware.LogRoller
}

const (
	// DefaultLogFilename is the default log filename.
	DefaultLogFilename = "query.log"
	// CommonLogFormat is the common log format.
	CommonLogFormat = `{remote} ` + CommonLogEmptyValue + ` [{when}] "{type} {class} {name} {proto} {>do} {>bufsize}" {rcode} {size} {duration}`
	// CommonLogEmptyValue is the common empty log value.
	CommonLogEmptyValue = "-"
	// CombinedLogFormat is the combined log format.
	CombinedLogFormat = CommonLogFormat + ` "{>opcode}"`
	// DefaultLogFormat is the default log format.
	DefaultLogFormat = CommonLogFormat
	DefaultEndpoint = "http://localhost:10000"
)
