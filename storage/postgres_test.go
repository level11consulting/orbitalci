package storage

import (
	"bitbucket.org/level11consulting/go-til/test"
	util "bitbucket.org/level11consulting/ocelot/common/testutil"
	"bitbucket.org/level11consulting/ocelot/models"

	"bytes"
	"testing"
	"time"
)

func TestPostgresStorage_AddSumStart(t *testing.T) {
	util.BuildServerHack(t)
	cleanup, pw, port := CreateTestPgDatabase(t)
	defer cleanup(t)
	pg := NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	pg.Connect()
	defer PostgresTeardown(t, pg.db)
	const shortForm = "2006-01-02 15:04:05"
	buildTime, err := time.Parse(shortForm,"2018-01-14 18:38:59")
	if err != nil {
		t.Error(err)
	}
	model := &models.BuildSummary{
		Hash: "123",
		Failed: false,
		BuildTime: buildTime,
		Account: "testAccount",
		BuildDuration: 23.232,
		Repo: "testRepo",
		Branch: "aBranch",
	}
	id, err := pg.AddSumStart(model.Hash, model.Account, model.Repo, model.Branch)
	if err != nil {
		t.Error(err)
	}
	t.Log("id ", id)
	sumaries, err := pg.RetrieveSum("123")
	if err != nil {
		t.Error(err)
	}
	sum := sumaries[0]
	if sum.Hash != "123" {
		t.Error(test.StrFormatErrors("hash", "123", sum.Hash))
	}
	// when first inserted, should be true
	if sum.Failed != true {
		t.Error(test.GenericStrFormatErrors("failed", true, sum.Failed))
	}
	//if sum.BuildTime != buildTime {
	//	t.Error(test.GenericStrFormatErrors("build start time", buildTime, sum.BuildTime))
	//}
	if sum.Account != "testAccount" {
		t.Error(test.StrFormatErrors("account", "testAccount", sum.Account))
	}
	if sum.Repo != "testRepo" {
		t.Error(test.StrFormatErrors("repo", "testRepo", sum.Repo))
	}
	if sum.Branch != "aBranch" {
		t.Error(test.StrFormatErrors("branch", "aBranch", sum.Branch))
	}
	err = pg.UpdateSum(model.Failed, model.BuildDuration, id)
	if err != nil {
		t.Error("could not update build summary: ", err)
	}
	//cleanup
	//_ = pg.db.QueryRow(`delete from build_summary where hash = 123`)
	sumaz, err := pg.RetrieveSum("123")
	if err != nil {
		t.Error(err)
	}
	suum := sumaz[0]
	if suum.BuildDuration != model.BuildDuration {
		t.Error(test.GenericStrFormatErrors("build duration", model.BuildDuration, suum.BuildDuration))
	}
	if suum.Failed != false {
		t.Error(test.GenericStrFormatErrors("failed", false, sum.Failed))
	}
}

func TestPostgresStorage_AddOut(t *testing.T) {
	util.BuildServerHack(t)
	pg, id, cleanup := insertDependentData(t)
	defer cleanup(t)
	defer PostgresTeardown(t, pg.db)
	txt := []byte("a;lsdkfjakl;sdjfakl;sdjfkl;asdj c389uro23ijrh8234¬˚å˙∆ßˆˆ…∂´¨¨;lsjkdafal;skdur23;klmnvxzic78r39q;lkmsndf")
	out := &models.BuildOutput{
		BuildId: id,
		Output: txt,
	}
	err := pg.AddOut(out)
	if err != nil {
		t.Error("could not add out: ", err)
	}
	retrieved, err := pg.RetrieveOut(id)
	if err != nil {
		t.Error("could not retrieve out: ", err)
	}
	if retrieved.BuildId != id {
		t.Error(test.GenericStrFormatErrors("build id", id, retrieved.BuildId))
	}
	if bytes.Compare(retrieved.Output, txt) != 0{
		t.Error(test.GenericStrFormatErrors("output", txt, retrieved.Output))
	}

}

func TestPostgresStorage_AddStageDetail(t *testing.T) {
	util.BuildServerHack(t)
	pg, id, cleanup := insertDependentData(t)
	defer cleanup(t)
	pg.Connect()
	defer PostgresTeardown(t, pg.db)
	const shortForm = "2006-01-02 15:04:05"
	startTime, _ := time.Parse(shortForm,"2018-01-14 18:38:59")
	stageMessage := []string{"wow I am amazing"}

	stageResult := &models.StageResult{
		BuildId: id,
		Error: "",
		StartTime: startTime,
		StageDuration: 100,
		Status: 1,
		Messages: stageMessage,
		Stage: "marianne",
	}
	err := pg.AddStageDetail(stageResult)
	t.Log(pg.db.Stats().OpenConnections)
	if err != nil {
		t.Error("could not add stage details", err)
	}

	stageResults, err := pg.RetrieveStageDetail(id)
	t.Log(pg.db.Stats().OpenConnections)
	if err != nil {
		t.Error("could not get stage details", err)
	}

	if len(stageResults) != 1 {
		t.Error(test.GenericStrFormatErrors("stage length", 1, len(stageResults)))
	}

	for _, stage := range stageResults {
		if stage.StageResultId != 1 {
			t.Error(test.GenericStrFormatErrors("postgres assigned stage result id", 1, stage.StageResultId))
		}
		if stage.BuildId != 1 {
			t.Error(test.GenericStrFormatErrors("test build id", 1, stage.BuildId))
		}
		if len(stage.Error) != 0 {
			t.Error(test.GenericStrFormatErrors("stage err length", 0, len(stage.Error)))
		}
		if stage.Stage != "marianne" {
			t.Error(test.GenericStrFormatErrors("stage name", "marianne", stage.Stage))
		}
		if len(stage.Messages) != len(stageMessage) || stage.Messages[0] != stageMessage[0] {
			t.Error(test.GenericStrFormatErrors("stage messages", stageMessage, stage.Messages))
		}
		if stage.StageDuration != 100 {
			t.Error(test.GenericStrFormatErrors("stage duration", 100, stage.Messages))
		}
	}
}

func TestPostgresStorage_Healthy(t *testing.T) {
	cleanup, pw, port := CreateTestPgDatabase(t)
	pg := NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	time.Sleep(4*time.Second)
	defer cleanup(t)
	if !pg.Healthy() {
		t.Error("postgres storage instance should return healthy, it isn't.")
	}
	cleanup(t)
	time.Sleep(2*time.Second)
	if pg.Healthy() {
		t.Error("postgres storage instance has been shut down, should return not healthy")
	}

}

func TestPostgresStorage_GetLastData(t *testing.T) {
	cleanup, pw, port := CreateTestPgDatabase(t)
	pg := NewPostgresStorage("postgres", pw, "localhost", port, "postgres")
	time.Sleep(4*time.Second)
	defer cleanup(t)
	_, hashes, err := pg.GetLastData("level11consulting/ocelot")
	if err != nil {
		t.Error(err)
	}
	if last, ok := hashes["master"]; !ok {
		t.Error("hash map should have master branch, it doesnlt")
		t.Log(hashes)
	} else if last != "6363a8a4ef13227218dc5c6d40e78ddfeb21b623"{
		t.Error(test.StrFormatErrors("master last hash", "6363a8a4ef13227218dc5c6d40e78ddfeb21b623", last))
	}
}

//todo: add cred tests for postgres!!!! plzplzplz