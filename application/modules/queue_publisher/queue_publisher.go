package queue_publisher

import (
	"devgit.kf.com.br/core/lib-queue/queue"
	"devgit.kf.com.br/core/lib-queue/queue_repository"
	"github.com/ivanmeca/timedEvent/application/modules/config"
	"github.com/ivanmeca/timedEvent/application/modules/logger"
	"os"
	"strconv"
	"sync"
)

var once sync.Once
var instance *queue_publisher

type queue_publisher struct {
	queueMap map[string]*queue.Queue
	mutex    sync.Mutex
}

func QueuePublisher() *queue_publisher {
	once.Do(func() {
		instance = &queue_publisher{queueMap: map[string]*queue.Queue{}}
		instance.init()
	})
	return instance
}

func (qp *queue_publisher) init() {
	AppLogger := logger.GetLogger()
	QueueConf := config.GetConfig().PublishQueue
	for _, qConf := range QueueConf {
		port, err := strconv.Atoi(qConf.ServerPort)
		if err != nil {
			AppLogger.ErrorPrintln("could not get queue port on queue " + qConf.QueueName)
			os.Exit(1)
		}
		params := queue_repository.NewQueueRepositoryParams(qConf.ServerUser, qConf.ServerPassword, qConf.ServerHost, port)
		params.SetVHost(qConf.ServerVHost)
		qr, err := queue_repository.NewQueueRepository(params)
		if err != nil {
			AppLogger.ErrorPrintln("could not init queue repository on queue " + qConf.QueueName)
			os.Exit(1)
		}
		queueName := qConf.QueueName
		qParam := queue.NewQueueParams(queueName)
		qParam.SetThreadLimit(200)
		q, err := qr.QueueDeclare(qParam, false)
		if err != nil {
			AppLogger.ErrorPrintln("could not declare queue " + queueName)
			os.Exit(1)
		}
		qp.queueMap[queueName] = q
	}
}

func (qp *queue_publisher) ValidateQueue(queueName string) bool {
	if _, ok := qp.queueMap[queueName]; ok {
		return true
	}
	return false
}

func (qp *queue_publisher) PublishInQueue(queueName string, data interface{}) bool {
	AppLogger := logger.GetLogger()
	qp.mutex.Lock()
	defer qp.mutex.Unlock()
	if val, ok := qp.queueMap[queueName]; ok {
		err := val.Publish(data)
		if err != nil {
			AppLogger.ErrorPrintln("could not publish on queue: " + err.Error())
			os.Exit(1)
			return false
		}
		return true
	}
	return false
}
