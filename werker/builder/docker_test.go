package builder

import (
	"bitbucket.org/level11consulting/go-til/net"
	pb "bitbucket.org/level11consulting/ocelot/protos"
	"bitbucket.org/level11consulting/ocelot/util/cred"
	"bitbucket.org/level11consulting/ocelot/util/integrations/dockr"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"

	"net/http"
	"os"
	"testing"
	"time"
)


// An extremely involved integration test with our local nexus repo and a docker image being brought up.
func TestDocker_RepoIntegrationSetup(t *testing.T) {
	// skip if is running in -short mode or if the nexus pw isn't set
	if testing.Short() {
		t.Skip("skipping due to -short flag being set")
	}
	password, ok := os.LookupEnv("NEXUS_ADMIN_PW")
	if !ok {
		t.Skip("skipping because $NEXUS_ADMIN_PW not set")
	}

	acctName := "test"
	projectName := "project"
	ctx := context.Background()
	docker, cleanupFunc := CreateLivingDockerContainer(t, ctx, "docker:18.02.0-ce")
	defer cleanupFunc(t)

	pull := []string{"/bin/sh", "-c", "docker pull docker.metaverse.l11.com/busybox:test_do_not_delete"}
	su := InitStageUtil("testing")

	time.Sleep(time.Second)
	testRemoteConfig, vaultListener, consulServer := cred.TestSetupVaultAndConsul(t)
	defer cred.TeardownVaultAndConsul(vaultListener, consulServer)

	// try to do a docker pull on private repo without creds written. this should fail.
	out := make(chan []byte, 10000)
	result := docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, pull, out)
	if result.Status != pb.StageResultVal_FAIL {
		data := <-out
		t.Error("pull from metaverse should fail if there are no creds to authenticate with. stdout: " , data)
	}
	// add in docker repo credentials,
	//cred.AddDockerRepoCreds(t, testRemoteConfig, "docker.metaverse.l11.com", password, "admin", acctName, projectName)

	// create config in ~/.docker directory w/ auth creds
	logout := make(chan[]byte, 10000)
	res := docker.IntegrationSetup(ctx, dockr.GetDockerConfig, docker.WriteDockerJson, "docker login", testRemoteConfig, acctName, su, []string{}, logout)
	if res.Status == pb.StageResultVal_FAIL {
		data := <- logout
		t.Error("stage failed! logout data: ", string(data))
	}

	// try to pull the image from metaverse again. this should now pass.
	logout = make(chan[]byte, 100000)
	res = docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, pull, logout)
	outByte := <- logout
	if res.Status == pb.StageResultVal_FAIL {
		t.Error("could not pull from metaverse docker! out: ", string(outByte))
	}
	muxi := mux.NewRouter()
	muxi.HandleFunc("/kubectl", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://storage.googleapis.com/kubernetes-release/release/v1.9.6/bin/linux/amd64/kubectl", 301)
	})

	n := net.InitNegroni("werker", muxi)
	go n.Run(":8888")

	result = docker.IntegrationSetup(ctx, func(config cred.CVRemoteConfig, acctName string)(string, error) {return "8888", nil}, docker.DownloadKubectl, "kubectl download", testRemoteConfig, acctName, su, []string{}, logout)
	outBytes := <-logout
	if result.Status == pb.StageResultVal_FAIL {
		t.Error("couldn't download kubectl! oh nuuuuuu! ", string(outBytes))
	}
	checkKube := []string{"/bin/sh", "-c", "command -v kubectl"}
	outd := make(chan []byte, 10000)
	result = docker.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, checkKube, outd)
	outb := <- outd
	if result.Status == pb.StageResultVal_FAIL {
		t.Error("kubectl not found! fail! ", string(outb))
	}
	t.Log(result.Status)
	t.Log(string(outb))
}

// test that in docker, can run the InstallPackageDeps to multiple image types
func TestDockerBasher_InstallPackageDeps(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping due to -short flag being set")
	}
	ctx := context.Background()
	alpine, cleanupFunc := CreateLivingDockerContainer(t, ctx, "alpine:latest")
	defer cleanupFunc(t)
	su := InitStageUtil("alpineTest")
	logout := make(chan[]byte, 10000)
	result := alpine.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, alpine.InstallPackageDeps(), logout)
	if result.Status == pb.StageResultVal_FAIL {
		t.Error("couldn't download deps! oh nuuu!")
	}
	t.Log(result.Status)
	t.Log(string(<-logout))
	testDeps := []string{"/bin/sh", "-c", "command -v openssl && command -v bash && command -v zip && command -v wget && command -v python"}
	result = alpine.Exec(ctx, su.GetStage(), su.GetStageLabel(), []string{}, testDeps, logout)
	if result.Status == pb.StageResultVal_FAIL {
		t.Error("deps not found! oh nuuu!")
	}
	t.Log(result.Status)
	t.Log(string(<-logout))



}
