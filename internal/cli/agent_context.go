// Copyright 2026 Jiahong Chen and contributors. Licensed under Apache-2.0. See LICENSE.
// Maintained in the World of Hyatt CLI repository.

package cli

import (
	"encoding/json"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// agentContextSchemaVersion is bumped on any breaking change to the JSON
// shape emitted by `agent-context`. Agents should check this before
// parsing. Shape at v3 adds kind-aware auth env var metadata.
const agentContextSchemaVersion = "3"

// agentContext is the structured description of this CLI consumed by AI
// agents. Inspired by Cloudflare's /cdn-cgi/explorer/api runtime endpoint
// (2026-04-13 Wrangler post): agents can introspect the live CLI without
// parsing --help or reading source.
type agentContext struct {
	SchemaVersion              string                 `json:"schema_version"`
	CLI                        agentContextCLI        `json:"cli"`
	Auth                       agentContextAuth       `json:"auth"`
	Discovery                  *agentContextDiscovery `json:"discovery,omitempty"`
	Commands                   []agentContextCommand  `json:"commands"`
	AvailableProfiles          []string               `json:"available_profiles"`
	FeedbackEndpointConfigured bool                   `json:"feedback_endpoint_configured"`
}

type agentContextCLI struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type agentContextAuth struct {
	Mode    string                   `json:"mode"`
	EnvVars []agentContextAuthEnvVar `json:"env_vars"`
}

type agentContextAuthEnvVar struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Required    bool   `json:"required"`
	Sensitive   bool   `json:"sensitive"`
	Description string `json:"description,omitempty"`
}

type agentContextDiscovery struct {
	Source            string   `json:"source"`
	TargetURL         string   `json:"target_url,omitempty"`
	EntryCount        int      `json:"entry_count,omitempty"`
	APIEntryCount     int      `json:"api_entry_count,omitempty"`
	Reachability      string   `json:"reachability,omitempty"`
	Protocols         []string `json:"protocols,omitempty"`
	AuthCandidates    []string `json:"auth_candidates,omitempty"`
	Protections       []string `json:"protections,omitempty"`
	GenerationHints   []string `json:"generation_hints,omitempty"`
	Warnings          []string `json:"warnings,omitempty"`
	CandidateCommands []string `json:"candidate_commands,omitempty"`
}

type agentContextCommand struct {
	Name        string                `json:"name"`
	Use         string                `json:"use,omitempty"`
	Short       string                `json:"short,omitempty"`
	Annotations map[string]string     `json:"annotations,omitempty"`
	Flags       []agentContextFlag    `json:"flags,omitempty"`
	Subcommands []agentContextCommand `json:"subcommands,omitempty"`
}

type agentContextFlag struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Usage   string `json:"usage,omitempty"`
	Default string `json:"default,omitempty"`
}

func newAgentContextCmd(rootCmd *cobra.Command) *cobra.Command {
	var pretty bool
	cmd := &cobra.Command{
		Use:         "agent-context",
		Short:       "Emit structured JSON describing this CLI for agents",
		Annotations: map[string]string{"mcp:read-only": "true"},
		Long: `Outputs a machine-readable description of commands, flags, and auth so
agents can introspect this CLI at runtime without parsing --help or
reading source. Schema is versioned via schema_version.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := buildAgentContext(rootCmd)
			enc := json.NewEncoder(os.Stdout)
			if pretty {
				enc.SetIndent("", "  ")
			}
			return enc.Encode(ctx)
		},
	}
	cmd.Flags().BoolVar(&pretty, "pretty", false, "indent JSON output for human reading")
	return cmd
}

func buildAgentContext(rootCmd *cobra.Command) agentContext {
	envVars := []agentContextAuthEnvVar{
		{
			Name:        "HYATT_COOKIES",
			Kind:        "per_call",
			Required:    false,
			Sensitive:   true,
			Description: "Optional raw Hyatt Cookie header for debugging direct HTTP transport. Normal live searches use browser-use.",
		},
		{
			Name:        "HYATT_TRANSPORT",
			Kind:        "runtime",
			Required:    false,
			Sensitive:   false,
			Description: "Optional override. Default is browser. Set to http/direct only when debugging raw HTTP.",
		},
		{
			Name:        "HYATT_BROWSER_SESSION",
			Kind:        "runtime",
			Required:    false,
			Sensitive:   false,
			Description: "Optional browser-use session name. Defaults to hyatt-cli.",
		},
		{
			Name:        "HYATT_BROWSER_PROFILE",
			Kind:        "runtime",
			Required:    false,
			Sensitive:   false,
			Description: "Optional browser-use profile name. Leave unset unless the local browser-use profile is known to work.",
		},
	}
	authMode := "browser"
	if authMode == "" {
		authMode = "none"
	}
	profiles := ListProfileNames()
	if profiles == nil {
		profiles = []string{}
	}
	return agentContext{
		SchemaVersion: agentContextSchemaVersion,
		CLI: agentContextCLI{
			Name:        "hyatt-cli",
			Description: "Hyatt award availability as a scriptable, local, agent-readable CLI.",
			Version:     rootCmd.Version,
		},
		Auth: agentContextAuth{
			Mode:    authMode,
			EnvVars: envVars,
		},
		Discovery:                  buildAgentDiscoveryContext(),
		Commands:                   collectAgentCommands(rootCmd),
		AvailableProfiles:          profiles,
		FeedbackEndpointConfigured: FeedbackEndpointConfigured(),
	}
}

func buildAgentDiscoveryContext() *agentContextDiscovery {
	return &agentContextDiscovery{
		Source:        "traffic-analysis",
		TargetURL:     "https://www.hyatt.com/explore-hotels/rate-calendar",
		EntryCount:    2,
		APIEntryCount: 1,
		Reachability:  "browser-use live pages; raw HTTP commonly returns 403",
		Protocols: []string{
			"html-embedded-state (90% confidence)",
		},
		AuthCandidates: []string{},
		Protections: []string{
			"hyatt-browser-clearance (80% confidence)",
		},
		GenerationHints: []string{
			"Use browser transport first for Hyatt hotel metadata and rate-calendar endpoints.",
			"Add a hand-authored parser that extracts the JavaScript assignment window.STORE = {...}; from HTML and emits normalized availability rows.",
			"Treat direct HTTP 403/429 as expected; HYATT_TRANSPORT=http is for debugging, not the default agent path.",
		},
		Warnings: []string{
			"html-state-not-standard-json-script: The calendar payload is a JavaScript assignment, not script#__NEXT_DATA__; built-in embedded-json extraction may not parse it without hand code.",
		},
		CandidateCommands: []string{
			"calendar — Fetch and parse a Hyatt Points Calendar page for one hotel spirit code.",
			"scan — Repeat calendar fetches across multiple spirit codes and date windows to find points availability.",
		},
	}
}

// collectAgentCommands walks the cobra tree from the given command and
// returns its direct children (skipping the agent-context command itself
// to avoid self-reference). Each child is recursed into if it has
// subcommands. Flags are captured via VisitAll. Output is sorted by
// command name for stable diffs across regenerations.
//
// Cobra's Hidden flag suppresses listing in --help but does not gate
// agent discovery. Raw resource parents are Hidden so --help stays
// curated and the `api` browser populates; the agent-context surface
// must still enumerate them and their endpoints so agents can call any
// action a CLI user could.
func collectAgentCommands(c *cobra.Command) []agentContextCommand {
	children := c.Commands()
	sort.Slice(children, func(i, j int) bool { return children[i].Name() < children[j].Name() })

	out := make([]agentContextCommand, 0, len(children))
	for _, sub := range children {
		if sub.Name() == "agent-context" {
			continue
		}
		entry := agentContextCommand{
			Name:  sub.Name(),
			Use:   sub.Use,
			Short: sub.Short,
		}
		// Surface Cobra annotations (e.g., hyatt:endpoint, mcp:read-only) so
		// agents and the live-dogfood classifier can detect destructive-at-auth
		// endpoints without parsing source. Empty maps are stripped via
		// omitempty in the struct tag.
		if len(sub.Annotations) > 0 {
			entry.Annotations = make(map[string]string, len(sub.Annotations))
			for k, v := range sub.Annotations {
				entry.Annotations[k] = v
			}
		}
		sub.Flags().VisitAll(func(f *pflag.Flag) {
			entry.Flags = append(entry.Flags, agentContextFlag{
				Name:    f.Name,
				Type:    f.Value.Type(),
				Usage:   f.Usage,
				Default: f.DefValue,
			})
		})
		sort.Slice(entry.Flags, func(i, j int) bool {
			return entry.Flags[i].Name < entry.Flags[j].Name
		})
		if len(sub.Commands()) > 0 {
			entry.Subcommands = collectAgentCommands(sub)
		}
		out = append(out, entry)
	}
	return out
}
