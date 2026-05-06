package populate

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Southclaws/fault"
	"github.com/Southclaws/fault/ftag"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
	"github.com/siemens/wfx/cmd/wfx/cmd/config"
	"github.com/siemens/wfx/generated/api"
	"github.com/siemens/wfx/internal/handler/job/definition"
	"github.com/siemens/wfx/workflow/dau"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func run(cmd *cobra.Command, k *koanf.Koanf) error {
	ctx := cmd.Context()

	appConfig, err := config.NewAppConfig(cmd.Flags())
	if err != nil {
		return fault.Wrap(err)
	}

	storage, err := appConfig.InitStorage()
	if err != nil {
		return fault.Wrap(err)
	}

	wf := dau.DirectWorkflow()
	if _, err := storage.GetWorkflow(ctx, wf.Name); err != nil {
		switch ftag.Get(err) {
		case ftag.NotFound:
			if _, err := storage.CreateWorkflow(ctx, wf); err != nil {
				return fault.Wrap(err)
			}
			log.Info().Str("name", wf.Name).Msgf("Created workflow %q", wf.Name)
		default:
			return fault.Wrap(err)
		}
	} else {
		log.Info().Str("name", wf.Name).Msgf("Workflow %q already exists", wf.Name)
	}

	count := k.Int(flagCount)
	workers := k.Int(flagWorkers)
	log.Info().Int("count", count).Int("workers", workers).Msg("Populating storage with jobs")

	tags := []string{"EUROPE_WEST", "loadtest"}

	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(workers)
	for i := range count {
		g.Go(func() error {
			now := time.Now()
			clientID := fmt.Sprintf("LOAD-%d", i)
			state := wf.States[rand.Intn(len(wf.States))] // pick random state
			job := api.Job{
				ClientID: clientID,
				Workflow: wf,
				Mtime:    &now,
				Stime:    &now,
				Status: &api.JobStatus{
					ClientID: clientID,
					State:    state.Name,
				},
				Definition: map[string]any{},
				Tags:       &tags,
				History:    &[]api.History{},
			}
			job.Status.DefinitionHash = definition.Hash(&job)
			_, err := storage.CreateJob(ctx, &job)
			return fault.Wrap(err)
		})
	}
	if err := g.Wait(); err != nil {
		return fault.Wrap(err)
	}

	log.Info().Int("count", count).Msg("Finished populating storage with jobs")

	return nil
}
