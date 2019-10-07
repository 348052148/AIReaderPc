package worker

import (
	"time"
	"sync"
	"container/list"
)

type Task struct {
	Func func(task *Task)
	Parame map[string]interface{}
	Result interface{}
	WorkerCode int
	Wg *sync.WaitGroup
}

func NewTask(f func(task *Task)) *Task {
	return &Task{
		Func:f,
		Parame:make(map[string]interface{}),
		Wg:&sync.WaitGroup{},
	}
}

func (task *Task)Run()  {
	if task.Wg != nil {
		task.Wg.Add(1)
		defer task.Wg.Done()
	}
	task.Func(task)
}

func (task *Task)Wait()  {
	if task.Wg != nil {
		task.Wg.Wait()
	}
}

func (task *Task)ReTry()  {
	task.Func(task)
}

type Worker struct {
	TaskList *list.List
	TaskChan chan *Task
	ShutDownChan chan struct{}
	WorkerCount int
	Wg *sync.WaitGroup
	TaskLoopFlag bool
	Lock *sync.Mutex
	ALock *sync.Mutex
}

func NewWorker() *Worker {
	workers := &Worker{
		TaskChan:make(chan *Task),
		ShutDownChan:make(chan struct{}),
		Wg:&sync.WaitGroup{},
		WorkerCount: 20,
		TaskList: list.New(),
		TaskLoopFlag: true,
		Lock: &sync.Mutex{},
		ALock:&sync.Mutex{},
	}
	go func() {
		for {
			select {
			case <-time.After(time.Millisecond * 100):
				workers.Lock.Lock()
				if workers.TaskLoopFlag {
					workers.TaskJoinLoop()
				}
				workers.Lock.Unlock()
			}
		}
	}()
	return workers
}

//添加任务
func (w *Worker)AddTask(task *Task)  {
	w.ALock.Lock()
	defer w.ALock.Unlock()
	w.TaskList.PushFront(task)
}
//弹出任务
func (w *Worker)PopTask() (*Task, bool) {
	w.ALock.Lock()
	defer w.ALock.Unlock()

	task := w.TaskList.Back()
	if task != nil {
		w.TaskList.Remove(task)
		return task.Value.(*Task), true
	}
	return nil, false
}

func (w *Worker)TaskJoinLoop()  {
	w.TaskLoopFlag = false
	defer func() {
		w.TaskLoopFlag = true
	}()
	for  {
		if task, ok := w.PopTask(); ok {
			w.TaskChan <- task
		} else {
			break
		}
	}

}

func (w *Worker)AddSyncTask(task *Task)  {
	task.Run()
}

func (w *Worker)AppendTask(task *Task)  {
	go func() {
		w.TaskChan <- task
	}()
}

//等待任务结束
func (w *Worker)WaitProcess()  {
	time.Sleep(time.Second * 5)
	w.Wg.Wait()
}

//关闭任务Workers
func (w *Worker)Shutdown()  {
	w.ShutDownChan <- struct{}{}
}

//任务调度
func (w *Worker)Run()  {
	w.Wg.Add(1)
	defer w.Wg.Done()
	for i:=0; i < w.WorkerCount; i++  {
		go func(workerCode int) {
			for  {
				select  {
				case task := <-w.TaskChan:
					w.Wg.Add(1)
					//设置workerCode
					task.WorkerCode = workerCode
					task.Run()
					//通知任务完成
					w.Wg.Done()
				case <-time.NewTimer(time.Second * 10).C:
				case <-w.ShutDownChan:
					break
				}
			}
		}(i)
	}
}