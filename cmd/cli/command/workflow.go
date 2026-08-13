package command

import (
	"fmt"

	clid "go-api/cmd/cli/di"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	domainworkflowrun "go-api/internal/domain/workflowrun"
	"go-api/internal/infrastructure/config"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewWorkflowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow",
		Short: "Workflow operations",
	}

	cmd.AddCommand(newWorkflowRunCommand())
	return cmd
}

func newWorkflowRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run <workflowId>",
		Short: "Start a workflow run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workflowID, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid workflow id")
			}

			env := config.Load()
			db := config.ConnectDatabase(env)
			container := clid.NewContainer(db, env)

			run, err := container.StartWorkflowRunHandler.Handle(cmd.Context(), workflowruncmd.StartWorkflowRunCommand{
				WorkflowID:  workflowID,
				TriggeredBy: domainworkflowrun.TriggeredByCLI,
			})
			if err != nil {
				return err
			}

			fmt.Printf("workflow run started id=%s workflowId=%s status=%s\n", run.ID, run.WorkflowID, run.Status)
			return nil
		},
	}
}
