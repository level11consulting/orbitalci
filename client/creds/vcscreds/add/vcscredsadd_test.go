package buildcredsadd

import (
	"bytes"
	"context"
	"flag"
	"strings"
	"testing"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mitchellh/cli"
	"github.com/shankj3/go-til/test"
	"github.com/level11consulting/ocelot/client/commandhelper"
	"github.com/level11consulting/ocelot/common/testutil"
	models "github.com/level11consulting/ocelot/models/pb"
)

// testNew will return the bare minimum. flags and fileloc of yaml will have to be set after instantiation
// or be generated by new functions
func testNew(inputReaderData []byte) *cmd {
	ui := cli.NewMockUi()
	if len(inputReaderData) >= 0 {
		ui.InputReader = bytes.NewReader(inputReaderData)
	}
	c := &cmd{
		UI:     ui,
		config: commandhelper.NewTestClientConfig([]string{}),
	}
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.fileloc, "credfile-loc", "",
		"Location of yaml file containing creds to upload")
	return c
}

func Test_cmd_Run_Yaml(t *testing.T) {
	var input []byte
	ui := cli.NewMockUi()
	if len(input) >= 0 {
		ui.InputReader = bytes.NewReader(input)
	}
	c := &cmd{
		UI:     ui,
		config: commandhelper.NewTestClientConfig([]string{}),
	}
	c.flags = flag.NewFlagSet("", flag.ContinueOnError)
	c.flags.StringVar(&c.fileloc, "credfile-loc", "",
		"Location of yaml file containing creds to upload")

	ctx := context.Background()
	c.fileloc = "./test-fixtures/newcreds.yml"
	expectedCreds := &models.CredWrapper{
		Vcs: []*models.VCSCreds{
			{
				ClientId:     "fancy-frickin-identification",
				ClientSecret: "SHH-BE-QUIET-ITS-A-SECRET",
				TokenURL:     "https://ocelot.perf/site/oauth2/access_token",
				AcctName:     "lamb-shank",
				SubType:      models.SubCredType_BITBUCKET,
				SshFileLoc:   "THIS IS A TEST",
			},
		},
	}
	var args []string
	//c.runCredFileUpload(ctx)
	if exit := c.Run(args); exit != 0 {
		t.Log(ui.OutputWriter.String())
		t.Log(ui.ErrorWriter.String())
		t.Fatal("should return exit 0 because even though SSH Key path doesn't exist, it doesn't mean failure")
	}
	uploadErrMsg := strings.TrimSpace(ui.ErrorWriter.String())
	expectedSSHErrMsg := `Could not read file at THIS IS A TEST 
Error: open THIS IS A TEST: no such file or directory`
	if uploadErrMsg != expectedSSHErrMsg {
		t.Error(test.StrFormatErrors("ssh key file error ouput", expectedSSHErrMsg, uploadErrMsg))
	}

	actualCreds, err := c.config.Client.GetVCSCreds(ctx, &empty.Empty{})
	if err != nil {
		t.Fatal("could not get actual creds from fake guide ocelot client")
	}
	if !testutil.CompareCredWrappers(expectedCreds, actualCreds) {
		t.Error("expected creds mismatch\n expected: ", expectedCreds, "\n actual: ", actualCreds)
	}

}

func Test_cmd_Run_noYaml(t *testing.T) {
	input := []byte(`lamb-shank
bitbucket
fancy-frickin-identification
SHH-BE-QUIET-ITS-A-SECRET`)
	cmd := testNew(input)
	ctx := context.Background()
	expectedCreds := &models.CredWrapper{
		Vcs: []*models.VCSCreds{
			{
				ClientId:     "fancy-frickin-identification",
				ClientSecret: "SHH-BE-QUIET-ITS-A-SECRET",
				AcctName:     "lamb-shank",
				SubType:      models.SubCredType_BITBUCKET,
				SshFileLoc:   "THIS IS A TEST",
			},
		},
	}

	var args []string
	exit := cmd.Run(args)
	if exit != 0 {
		t.Error("should return exit code 0, got ", exit)
	}
	sentCreds, err := cmd.config.Client.GetVCSCreds(ctx, &empty.Empty{})
	if err != nil {
		t.Fatal("could not get actual creds from fake guide ocelot client")
	}
	if !testutil.CompareCredWrappers(expectedCreds, sentCreds) {
		t.Error("expected creds mismatch\n expected: ", expectedCreds, "\n actual: ", sentCreds)
	}
}
