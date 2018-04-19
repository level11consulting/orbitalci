package listener


import (
	ocelog "bitbucket.org/level11consulting/go-til/log"
	"bitbucket.org/level11consulting/ocelot/build"
	"bitbucket.org/level11consulting/ocelot/build/basher"
	bldr "bitbucket.org/level11consulting/ocelot/build/builder"
	"bitbucket.org/level11consulting/ocelot/build/launcher"
	"bitbucket.org/level11consulting/ocelot/build/valet"
	"bitbucket.org/level11consulting/ocelot/common/credentials"
	"bitbucket.org/level11consulting/ocelot/models"
	"bitbucket.org/level11consulting/ocelot/models/pb"
	"bitbucket.org/level11consulting/ocelot/storage"
	"github.com/golang/protobuf/proto"

	//"runtime/debug"
	"fmt"
)

type WorkerMsgHandler struct {
	*models.WerkerFacts
	Topic           string
	Type        models.WerkType
	infochan        chan []byte
	StreamChan   chan *models.Transport
	BuildCtxChan chan *models.BuildContext
	Basher          *basher.Basher
	Store           storage.OcelotStorage
	BuildValet   *valet.Valet
	RemoteConfig  credentials.CVRemoteConfig

}

func NewWorkerMsgHandler(topic string, facts *models.WerkerFacts, b *basher.Basher, st storage.OcelotStorage, bv *valet.Valet, rc credentials.CVRemoteConfig, tunnel chan *models.Transport, buildChan chan *models.BuildContext) *WorkerMsgHandler {
	return &WorkerMsgHandler{
		Topic: 		   topic,
		Basher: 	   b,
		Store: 		   st,
		BuildValet:	   bv,
		StreamChan:   tunnel,
		BuildCtxChan: buildChan,
		RemoteConfig: rc,
		WerkerFacts:  facts,
	}
}

// UnmarshalAndProcess is called by the nsq consumer to handle the build message
// It uses two channels to communicate with nsqpb, done and finish.
// the done channel is just sent at the end and is used in nsqpb to ensure that the queue is "Touch"ed at a
// set interval so that the message doesn't time out. The finish channel is for improper exits; ie panic recover
// or signal handling
// The nsqpb will call msg.Finish() when it receives on this channel.
func (w WorkerMsgHandler) UnmarshalAndProcess(msg []byte, done chan int, finish chan int) error {
	ocelog.Log().Debug("unmarshal-ing build obj and processing")
	werkerTask := &pb.WerkerTask{}
	if err := proto.Unmarshal(msg, werkerTask); err != nil {
		ocelog.IncludeErrField(err).Warning("unmarshal error")
		return err
	}
	if err := w.Store.StartBuild(werkerTask.Id); err != nil {
		ocelog.IncludeErrField(err).Error("couldn't log start of build, returning")
		return err
	}
	// channels get closed after the build finishes
	w.infochan = make(chan []byte)
	ocelog.Log().Debug(fmt.Sprintf("INFO CHANNEL IS!!!!!  %v     MSGHANDLER IS!!!! %#v", w.infochan, w))
	// set goroutine for watching for results and logging them (for rn)
	// cant add go watchForResults here bc can't call method on interface until it's been cast properly.
	//
	var builder build.Builder
	switch w.Type {
	case models.Docker:
		builder = bldr.NewDockerBuilder(w.Basher)
	default:
		builder = bldr.NewDockerBuilder(w.Basher)
	}
	launch := launcher.NewLauncher(w.WerkerFacts, w.RemoteConfig, w.StreamChan, w.BuildCtxChan, w.Basher, w.Store, w.BuildValet)
	launch.MakeItSo(werkerTask, builder, finish, done)
	return nil
}
