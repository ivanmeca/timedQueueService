package timer_control

import (
	"context"
	"fmt"
	"github.com/ivanmeca/timedEvent/application/modules/database/collection_managment"
	"github.com/ivanmeca/timedEvent/application/modules/database/data_types"
	"github.com/ivanmeca/timedEvent/application/modules/logger"
	"github.com/ivanmeca/timedEvent/application/modules/queue_publisher"
	"sync"
	"time"
)

const TimerControlUnit = time.Millisecond

type TimerControl struct {
	expirationTime time.Duration
	controlTime    time.Duration
	list           *sync.Map
	logger         *logger.StdLogger
}

func NewTimerControl(controlTime int, expirationTime int, list *sync.Map) *TimerControl {
	return &TimerControl{list: list, controlTime: time.Duration(controlTime), expirationTime: time.Duration(expirationTime)}
}

func (tc *TimerControl) Run(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				tc.processList()
			}
		}
	}()
	tc.logger = logger.GetLogger()
}

func (tc *TimerControl) processList() {
	time.Sleep(tc.controlTime * TimerControlUnit)
	horaAtual := time.Now().UTC()
	tc.logger.DebugPrintln("Timer control:" + horaAtual.Format("2006-01-02 15:04:05Z"))
	tc.list.Range(func(key interface{}, value interface{}) bool {
		if event, ok := value.(data_types.EventMapper); ok {
			timeDiffInSecond := horaAtual.Sub(event.PublishDate)
			timeDiffInSecond /= TimerControlUnit
			tc.logger.DebugPrintln(fmt.Sprintf("Actual time: %s, event time: %s, timeDiff: %d", horaAtual.Format("2006-01-02 15:04:05Z"), event.PublishDate.Format("2006-01-02 15:04:05Z"), timeDiffInSecond))
			if timeDiffInSecond > tc.expirationTime {
				_, err := collection_managment.NewEventCollection().DeleteItem([]string{event.EventID})
				if err != nil {
					tc.logger.NoticePrintln("falha ao excluir ID: " + event.EventID)
				}
				tc.list.Delete(key)
				tc.logger.DebugPrintln("ID excluido: " + event.EventID)
			} else {
				if timeDiffInSecond >= 0 {
					data, err := collection_managment.NewEventCollection().ReadItem(event.EventID)
					if err != nil {
						tc.logger.ErrorPrintln("event check fail: " + err.Error())
						tc.list.Delete(key)
						return true
					}
					if data.ArangoRev == event.EventRevision {
						tc.logger.DebugPrintln("Publicar ID " + event.EventID)
						var dataToPublish interface{}
						if event.Event.PublishType == data_types.DataOnly {
							dataToPublish = event.Event.CloudEvent.Data
						} else {
							dataToPublish = event.Event.CloudEvent
						}
						go func() {
							if queue_publisher.QueuePublisher().PublishInQueue(data.PublishQueue, dataToPublish) {
								_, err := collection_managment.NewEventCollection().DeleteItem([]string{event.EventID})
								if err != nil {
									tc.logger.NoticePrintln("falha ao excluir ID: " + event.EventID)
								}
								tc.list.Delete(key)
								tc.logger.DebugPrintln("ID excluido: " + event.EventID)
							}
						}()
					} else {
						tc.list.Delete(key)
					}
				}
			}
		}
		return true
	})
}
