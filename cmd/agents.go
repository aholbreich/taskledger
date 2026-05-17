package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const agentsSnippet = "## TaskLedger Workflow\n" +
	"\n" +
	"This repository uses TaskLedger (`tl`) for local task coordination between humans and agents.\n" +
	"\n" +
	"When starting work:\n" +
	"\n" +
	"1. - Run `tl ready --json` to find tasks that are open, unblocked, and unclaimed.\n" +
	"   - Or `tl show <task-id>` when asked to do a particular task.\n" +
	"2. Claim one task before editing files:\n" +
	"   `tl claim <task-id> --actor <your-agent-name>`\n" +
	"3. Inspect the task details:\n" +
	"   `tl show <task-id>`\n" +
	"4. Do the work. Record important context, decisions, blockers, or handoff notes:\n" +
	"   `tl note <task-id> --actor <your-agent-name> -m \"...\"`\n" +
	"5. When the task is complete, close it:\n" +
	"   `tl close <task-id> --actor <your-agent-name>`\n" +
	"\n" +
	"Rules:\n" +
	"\n" +
	"- Do **not** work on a task claimed by another active actor unless explicitly told.\n" +
	"- Prefer tasks from `tl ready`; blocked, pending, done, cancelled, or actively claimed tasks are not ready.\n" +
	"- Leave notes for partial progress, failed approaches, decisions, and handoffs.\n" +
	"- Do **not** edit `.taskledger/events.jsonl` manually.\n" +
	"- Set `TL_ACTOR` when possible so commands can resolve your identity consistently.\n" +
	"- Ask before editing `AGENTS.md` or other project instruction files.\n" +
	"- If `.taskledger/` is missing, ask the human whether to run `tl init`.\n"

func newAgentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "agents",
		Short: "Print recommended AGENTS.md instructions",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprint(cmd.OutOrStdout(), agentsSnippet)
			return err
		},
	}
}
