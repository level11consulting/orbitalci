package builder

import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/integrations/nexus"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bufio"
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/mitchellh/go-homedir"
	"io"
	"strings"
	"fmt"
)

type Docker struct{
	Log	io.ReadCloser
	ContainerId	string
	DockerClient *client.Client
	*Basher
}

func NewDockerBuilder(b *Basher) Builder {
	return &Docker{nil, "", nil, b}
}

func (d *Docker) Setup(logout chan []byte, werk *pb.WerkerTask, rc cred.CVRemoteConfig) *Result {
	su := InitStageUtil("setup")

	logout <- []byte(su.GetStageLabel() + "Setting up...")

	ctx := context.Background()
	cli, err := client.NewEnvClient()
	d.DockerClient = cli

	if err != nil {
		return &Result{
			Stage:  su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}

	imageName := werk.BuildConf.Image

	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		ocelog.IncludeErrField(err).Error("couldn't pull image!")
		return &Result{
			Stage:  su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}
	//var byt []byte
	//buf := bufio.NewReader(out)
	//buf.Read(byt)
	//fmt.Println(string(byt))
	defer out.Close()

	bufReader := bufio.NewReader(out)
	d.writeToInfo(su.GetStageLabel(), bufReader, logout)

	logout <- []byte(su.GetStageLabel() + "Creating container...")

	//container configurations
	containerConfig := &container.Config{
		Image: imageName,
		Env: werk.BuildConf.Env,
		Cmd: d.DownloadCodebase(werk),
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:true,
		Tty:true,
	}

	homeDirectory, _ := homedir.Expand("~/.ocelot")
	// todo: render settingsxml if necessary
	//host config binds are mount points
	hostConfig := &container.HostConfig{
		//TODO: have it be overridable via env variable
		Binds: []string{ homeDirectory + ":/.ocelot", "/var/run/docker.sock:/var/run/docker.sock"},
		NetworkMode: "host",
	}

	resp, err := cli.ContainerCreate(ctx, containerConfig , hostConfig, nil, "")

	if err != nil {
		return &Result{
			Stage:  su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}

	for _, warning := range resp.Warnings {
		logout <- []byte(warning)
	}

	logout <- []byte(su.GetStageLabel() + "Container created with ID " + resp.ID)

	d.ContainerId = resp.ID
	ocelog.Log().Debug("starting up container")
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return &Result{
			Stage:  su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}

	logout <- []byte(su.GetStageLabel()  + "Container " + resp.ID + " started")

	//since container is created in setup, log tailing via container is also kicked off in setup
	containerLog, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow: true,
	})

	if err != nil {
		return &Result{
			Stage: su.GetStage(),
			Status: FAIL,
			Error:  err,
		}
	}

	d.Log = containerLog
	bufReader = bufio.NewReader(containerLog)
	d.writeToInfo(su.GetStageLabel() , bufReader, logout)
	if settingsXML, err := nexus.GetSettingsXml(rc, strings.Split(werk.FullName, "/")[0]); err != nil {
		_, ok := err.(*nexus.NoCreds)
		if !ok {
			return &Result{
				Stage: su.GetStage(),
				Status: FAIL,
				Error:  err,
			}
		}
	} else {
		ocelog.Log().Debug("writing maven settings.xml")
		result := d.Exec(su.GetStage(), su.GetStageLabel(), []string{}, d.WriteMavenSettingsXml(settingsXML), logout)
		//d.writeToInfo(su.GetStageLabel() , bufReader, logout)
		return result
	}
	return &Result{
		Stage:  su.GetStage(),
		Status: PASS,
		Error:  nil,
	}
}

func (d *Docker) Cleanup(logout chan []byte) {
	su := InitStageUtil("cleanup")
	logout <- []byte(su.GetStageLabel() + "Performing build cleanup...")

	//TODO: review, should we be creating new contexts for every stage?
	cleanupCtx := context.Background()
	if d.Log != nil {
		d.Log.Close()
	}
	if err := d.DockerClient.ContainerKill(cleanupCtx, d.ContainerId, "SIGKILL"); err != nil {
		ocelog.IncludeErrField(err).WithField("containerId", d.ContainerId).Error("couldn't kill")
	} else {
		if err := d.DockerClient.ContainerRemove(cleanupCtx, d.ContainerId, types.ContainerRemoveOptions{}); err != nil {
			ocelog.IncludeErrField(err).WithField("containerId", d.ContainerId).Error("couldn't rm")
		}
	}
	d.DockerClient.Close()
}


func (d *Docker) Execute(stage *pb.Stage, logout chan []byte, commitHash string) *Result {
	if len(d.ContainerId) == 0 {
		return &Result {
			Stage: stage.Name,
			Status: FAIL,
			Error: errors.New("no container exists, setup before executing"),
		}
	}

	su := InitStageUtil(stage.Name)
	return d.Exec(su.GetStage(), su.GetStageLabel(), stage.Env, d.BuildScript(stage.Script, commitHash), logout)
}

//uses the repo creds from admin to store artifact(s)
func (d *Docker) SaveArtifact(logout chan []byte, task *pb.WerkerTask) *Result {
	su := &StageUtil{
		Stage: "SaveArtifact",
		StageLabel: "SAVE_ARTIFACT | ",
	}

	logout <- []byte(su.GetStageLabel() + "Saving artifact...")

	if len(d.ContainerId) == 0 {
		return &Result {
			Stage: su.GetStage(),
			Status: FAIL,
			Error: errors.New("no container exists, setup before executing"),
		}
	}

	//check if build tool if set to maven (cause that's the only thing that we use to push to nexus right now)
	if strings.Compare(task.BuildConf.BuildTool, "maven") != 0 {
		logout <- []byte(fmt.Sprintf(su.GetStageLabel() + "build tool %s not part of accepted values: %s...", task.BuildConf.BuildTool, "maven"))
		return &Result {
			Stage: su.GetStage(),
			Status: FAIL,
			Error: errors.New(fmt.Sprintf("build tool %s not part of accepted values: %s...", task.BuildConf.BuildTool, "maven")),
		}
	}

	//TODO: check if nexus creds exist

	return d.Exec(su.GetStage(), su.GetStageLabel(), nil, d.PushToNexus(task.CheckoutHash), logout)
}


func (d *Docker) Exec(currStage string, currStageStr string, env []string, cmds []string, logout chan []byte) *Result {
	ctx := context.Background()

	resp, err := d.DockerClient.ContainerExecCreate(ctx, d.ContainerId, types.ExecConfig{
		Tty: true,
		AttachStdin: true,
		AttachStderr: true,
		AttachStdout: true,
		Env: env,
		Cmd: cmds,
	})

	if err != nil {
		return &Result{
			Stage:  currStage,
			Status: FAIL,
			Error:  err,
		}
	}

	attachedExec, err := d.DockerClient.ContainerExecAttach(ctx, resp.ID, types.ExecConfig{
		Tty: true,
		AttachStdin: true,
		AttachStderr: true,
		AttachStdout: true,
		Env: env,
		Cmd: cmds,
	})

	defer attachedExec.Conn.Close()

	d.writeToInfo(currStageStr, attachedExec.Reader, logout)
	inspector, err := d.DockerClient.ContainerExecInspect(ctx, resp.ID)
	if inspector.ExitCode != 0 {
		return &Result{
			Stage: currStage,
			Status: FAIL,
			Error: nil,
			Messages: []string{"exit code was not zero"},
		}
	}

	if err != nil {
		return &Result{
			Stage:  currStage,
			Status: FAIL,
			Error:  err,
		}
	}
	return &Result{
		Stage:  currStage,
		Status: PASS,
		Error:  nil,
	}
}


func (d *Docker) writeToInfo(stage string, rd *bufio.Reader, infochan chan []byte) {
	scanner := bufio.NewScanner(rd)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		str := string(scanner.Bytes())
		infochan <- []byte(stage + str)
		//our setup script will echo this to stdout, telling us script is finished downloading. This is HACK for keeping container alive
		if strings.Contains(str, "Ocelot has finished with downloading source code") {
			ocelog.Log().Debug("finished with source code, returning out of writeToInfo")
			return
		}
	}
	ocelog.Log().Debug("finished writing to channel for stage ", stage)
	if err := scanner.Err(); err != nil {
		ocelog.IncludeErrField(err).Error("error outputing to info channel!")
		infochan <- []byte("OCELOT | BY THE WAY SOMETHING WENT WRONG SCANNING STAGE INPUT")
	}
}
