package admin

import (

	"github.com/shankj3/go-til/deserialize"
	"github.com/shankj3/go-til/nsqpb"

	"github.com/level11consulting/ocelot/build/helpers/buildscript/validate"
	"github.com/level11consulting/ocelot/client/buildconfigvalidator"
	"github.com/level11consulting/ocelot/models/pb"
	"github.com/level11consulting/ocelot/server/config"
	"github.com/level11consulting/ocelot/storage"
	"github.com/level11consulting/ocelot/router/admin/anycred"
)

//this is our grpc server, it responds to client requests
type OcelotServerAPI struct {
	anycred.AnyCredAPI // This is a hack. Revisit once stable
	BuildAPI
	PollScheduleAPI
	RepoInterfaceAPI
	StatusInterfaceAPI
	SecretInterfaceAPI
}

func NewGuideOcelotServer(config config.CVRemoteConfig, d *deserialize.Deserializer, adminV *validate.AdminValidator, repoV *validate.RepoValidator, storage storage.OcelotStorage, hhBaseUrl string) pb.GuideOcelotServer {

	anyCredAPI := anycred.AnyCredAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	buildAPI := BuildAPI {
		Storage:        storage,	
		RemoteConfig:   config,
		Deserializer:   d,
		Producer:       nsqpb.GetInitProducer(),
		OcyValidator:   buildconfigvalidator.GetOcelotValidator(),
	}

	appleDevSecretAPI := AppleDevSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	artifactRepoSecretAPI := ArtifactRepoSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
		RepoValidator:  repoV,
	}

	genericSecretAPI := GenericSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	kubernetesSecretAPI := KubernetesSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	notifierSecretAPI := NotifierSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	pollScheduleAPI := PollScheduleAPI {
		Storage:        storage,	
		RemoteConfig:   config,
		Producer:       nsqpb.GetInitProducer(),
	}

	repoInterfaceAPI := RepoInterfaceAPI {
		RemoteConfig:   config,
		Storage:        storage,	
		HhBaseUrl:      hhBaseUrl,
	}

	sshSecretAPI := SshSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	statusInterfaceAPI := StatusInterfaceAPI {
		Storage:        storage,	
		RemoteConfig:   config,
	}

	vcsSecretAPI := VcsSecretAPI {
		Storage:        storage,	
		RemoteConfig:   config,
		AdminValidator: adminV,
	}

	secretInterfaceAPI := SecretInterfaceAPI {
		AppleDevSecretAPI: appleDevSecretAPI,
		ArtifactRepoSecretAPI: artifactRepoSecretAPI,
		GenericSecretAPI: genericSecretAPI,
		KubernetesSecretAPI: kubernetesSecretAPI,
		NotifierSecretAPI: notifierSecretAPI,
		SshSecretAPI: sshSecretAPI,
		VcsSecretAPI: vcsSecretAPI,
	}

	return &OcelotServerAPI{ 
		anyCredAPI,
		buildAPI,
		pollScheduleAPI,
		repoInterfaceAPI,
		statusInterfaceAPI,
		secretInterfaceAPI,
	}
}
