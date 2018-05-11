package launcher

import (
	"strings"

	"github.com/shankj3/ocelot/build/notifiers"
	"github.com/shankj3/ocelot/build/notifiers/slack"
	"github.com/shankj3/ocelot/models"
	"github.com/shankj3/ocelot/models/pb"

)

func getNotifiers() []notifiers.Notifier {
	return []notifiers.Notifier{slack.Create()}
}

// doNotifications will notify everything you want it to. should be called at the end of a build
func (w *launcher) doNotifications(werk *pb.WerkerTask) error {
	accountName := strings.Split(werk.FullName, "/")[0]
	notifys := getNotifiers()
	for _, notify := range notifys {
		if !notify.IsRelevant(werk.BuildConf) {
			continue
		}
		credz, err := w.RemoteConf.GetCredsBySubTypeAndAcct(w.Store, notify.SubType(), accountName, false)
		if err != nil {
			return err
		}
		stageResults, err := w.Store.RetrieveStageDetail(werk.Id)
		if err != nil {
			return err
		}
		buildSum, err := w.Store.RetrieveSumByBuildId(werk.Id)
		if err != nil {
			return err
		}
		fullResult := models.ParseStagesByBuildId(buildSum, stageResults)

		err = notify.RunIntegration(credz, fullResult, werk.BuildConf.Notify)
		if err != nil {
			return err
		}
	}
	return nil
}