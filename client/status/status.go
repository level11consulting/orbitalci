package status

import (
	"bitbucket.org/level11consulting/ocelot/admin/models"
	"bitbucket.org/level11consulting/ocelot/client/commandhelper"
	"context"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"time"
)

const synopsis = "show status of specific acctname/repo, repo or hash"
const help = `
Usage: ocelot status 
	-hash <hash> [optional] if specified, this is the one that takes precendence over other arguments
	-acct-repo <acctname/repo> [optional] if specified, takes precedence over -repo argument
	-repo <repo> [optional] returns status of all repos starting with value
`

func New(ui cli.Ui) *cmd {
	// suppress ui here because there's an ordering to status and the error messages that come stock
	// with OcyHelper may be confusing
	c := &cmd{UI: ui, config: commandhelper.Config, OcyHelper: &commandhelper.OcyHelper{SuppressUI: true,}}
	c.init()
	return c
}

type cmd struct {
	UI cli.Ui
	flags   *flag.FlagSet
	config *commandhelper.ClientConfig

	*commandhelper.OcyHelper
}


func (c *cmd) GetClient() models.GuideOcelotClient {
	return c.config.Client
}

func (c *cmd) GetUI() cli.Ui {
	return c.UI
}

func (c *cmd) GetConfig() *commandhelper.ClientConfig {
	return c.config
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return help
}

func (c *cmd) init() {
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	//we accept all 3 flags, but prioritize output in the following order: hash, acct-repo, acct
	c.flags.StringVar(&c.OcyHelper.Hash, "hash", "ERROR", "[optional]  <hash> to display build status")
	c.flags.StringVar(&c.OcyHelper.Repo, "repo", "ERROR", "[optional]  <repo> to display build status")
	c.flags.StringVar(&c.OcyHelper.AcctRepo, "acct-repo", "ERROR", "[optional]  <account>/<repo> to display build status")
}

func (c *cmd) writeStatusErr(err error) {
	status, ok := grpcStatus.FromError(err)
	// if we can't parse the status, just return the shitty error.
	if !ok {
		c.UI.Error(err.Error())
	}
	if status.Code() == codes.NotFound {
		c.UI.Error(fmt.Sprintf("Status for %s was not found in the database. It may have not been processed yet.", c.OcyHelper.Hash))
	} else {
		// here we should post to admin
		c.UI.Error("Error retrieving status, message: " + status.Message())
	}
}

func (c *cmd) Run(args []string) int {
	var statuses *models.Status
	var err error
	if err := c.flags.Parse(args); err != nil {
		return 1
	}
	// if nothing is set, attempt to detect hash
	if c.OcyHelper.AcctRepo == "ERROR" && c.OcyHelper.Repo == "ERROR" && c.OcyHelper.Hash == "ERROR" {
		if err := c.OcyHelper.DetectHash(c); err != nil {
			commandhelper.Debuggit(c, err.Error())
			c.UI.Error("You must either be in the repository you want to track, one of the following flags must be set: -acct-repo, -repo, -hash. see --help")
			return 1
		}
	}

	ctx := context.Background()
	if err := commandhelper.CheckConnection(c, ctx); err != nil {
		return 1
	}

	// always respect hash first
	if c.OcyHelper.Hash != "ERROR" && len(c.OcyHelper.Hash) > 0 {
		query := &models.StatusQuery{
			Hash: c.OcyHelper.Hash ,
		}
		statuses, err = c.GetClient().GetStatus(ctx, query)
		if err != nil {
			c.writeStatusErr(err)
			return 1
		}
		goto STATUS_FOUND
		return 0
	}

	//respect acct-repo next
	if c.OcyHelper.AcctRepo != "ERROR" {
		if err := c.OcyHelper.SplitAndSetAcctRepo(c); err != nil {
			return 1
		}

		query := &models.StatusQuery{
			AcctName: c.OcyHelper.Account,
			RepoName: c.OcyHelper.Repo,
		}
		statuses, err = c.GetClient().GetStatus(ctx, query)
		if err != nil {
			c.writeStatusErr(err)
			return 1
		}
		goto STATUS_FOUND
	}

	//repo is last
	if c.OcyHelper.Repo != "ERROR" {
		query := &models.StatusQuery{
			PartialRepo: c.OcyHelper.Repo,
		}
		statuses, err = c.GetClient().GetStatus(ctx, query)
		if err != nil {
			c.writeStatusErr(err)
			return 1
		}
		goto STATUS_FOUND
		return 0
	}
	return 0
STATUS_FOUND:
	queued := statuses.BuildSum.BuildTime.Seconds == 0 && statuses.BuildSum.BuildTime.Nanos == 0
	buildStarted := statuses.BuildSum.BuildTime.Seconds > 0 && statuses.IsInConsul
	finished := !statuses.IsInConsul && statuses.BuildSum.BuildTime.Seconds > 0
	commandhelper.Debuggit(c, fmt.Sprintf("finished is %v, buildStarted is %v, queued is %v, buildTime is %v", finished, buildStarted, queued, time.Unix(statuses.BuildSum.BuildTime.Seconds, 0)))
	//statuses.BuildSum.QueueTime time.Unix(0,0)
	stageStatus, color, statuss := commandhelper.PrintStatusStages(commandhelper.GetStatus(queued, buildStarted, finished), statuses)
	buildStatus := commandhelper.PrintStatusOverview(color, statuses.BuildSum.Account, statuses.BuildSum.Repo, statuses.BuildSum.Hash, statuss)
	c.UI.Output(buildStatus + stageStatus)
	return 0
}