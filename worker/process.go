package worker

import (
	"sync"
)

//p.addTask().join(task).join.wait()

type WorkerProcess struct {
	Worker *Worker
	ResultChan chan interface{}
	Wg *sync.WaitGroup
}

func NewWorkerProcess() WorkerProcess {
	w := NewWorker()
	return WorkerProcess{
		Worker:w,
		ResultChan:make(chan interface{}),
		Wg:&sync.WaitGroup{},
	}
}

//分组
func (wp WorkerProcess)Group(fn func(wp WorkerProcess) []<-chan interface{}) WorkerProcess  {
	wp.ResultChan = fanIn(nil, fn(wp)...)
	return wp
}

func fanIn(done chan interface{}, inStreams ...<-chan interface{}) chan interface{} {
	outStream := make(chan interface{})
	wg := &sync.WaitGroup{}
	for _,inStream := range inStreams {
		wg.Add(1)
		go func(in <- chan interface{}) {
			defer wg.Done()
			for v := range in{
				select {
				case outStream <- v:
				case <-done:
					return
				}
			}
		}(inStream)
	}
	//
	go func() {
		wg.Wait()
		close(outStream)
	}()
	return outStream
}

func (wp WorkerProcess)AddTask(task *Task) WorkerProcess {
	wp.Worker.AppendTask(task)
	return wp
}

func (wp WorkerProcess)Join(f func(task *Task)) WorkerProcess {
	t := NewTask(f)
	t.Result = make(chan interface{})
	t.Parame["p"] = wp.ResultChan
	wp.Worker.AddTask(t)
	wp.ResultChan = t.Result.(chan interface{})
	return wp
}

func (wp WorkerProcess)Run() WorkerProcess {
	wp.Worker.Run()
	return wp
}
func (wp WorkerProcess)Wait()  {
	wp.Worker.WaitProcess()
}
