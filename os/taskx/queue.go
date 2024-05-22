package taskx

import (
	"sync"
)

// 任务处理调度器
type QueueScheduler struct {
	mutex *sync.Mutex
	Head  *QueueTask            // 首个任务
	Tail  *QueueTask            // 末尾任务
	Tasks map[string]*QueueTask // 所有任务
}

// 对列任务
type QueueTask struct {
	name string       // 任务名
	fn   func() error // 当前任务
	prev *QueueTask   // 指向上·一个任务
	next *QueueTask   // 指向下一个任务
}

func (t *QueueTask) HasNext() bool {
	return t.next != nil
}

func (t *QueueTask) Name() string {
	return t.name
}

func Queue() *QueueScheduler {
	return &QueueScheduler{
		mutex: new(sync.Mutex),
		Tasks: make(map[string]*QueueTask),
	}
}

// 执行
func (q *QueueScheduler) Execute() (err error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	for q.Head != nil {
		if err = q.Head.fn(); err != nil {
			break
		}
		delete(q.Tasks, q.Head.name)
		q.Head = q.Head.next
	}
	if len(q.Tasks) == 0 {
		q.Head = nil
		q.Tail = nil
	}
	return
}

// 新增（默认尾插）
func (q *QueueScheduler) Add(name string, fn func() error) {
	if name != "" && fn != nil {
		q.AddTail(name, fn)
	}
}

// 尾插（当前新增任务添加到队列末尾）
func (q *QueueScheduler) AddTail(name string, fn func() error) {
	if name != "" && fn != nil {
		q.mutex.Lock()
		defer q.mutex.Unlock()
		if _, exist := q.Tasks[name]; !exist {
			var add = &QueueTask{name: name, fn: fn}
			if tail := q.Tail; tail != nil {
				add.prev = tail
				tail.next = add
				q.Tail = add
			} else {
				q.Head = add
				q.Tail = add
			}
			q.Tasks[name] = add
		}
	}
}

// 头插（当前新增任务添加到队列首位）
func (q *QueueScheduler) AddHead(name string, task func() error) {
	if name != "" && task != nil {
		q.mutex.Lock()
		defer q.mutex.Unlock()
		if _, exist := q.Tasks[name]; !exist {
			var add = &QueueTask{name: name, fn: task}
			if head := q.Head; head != nil {
				head.prev = add
				add.next = head
				q.Head = add
			} else {
				q.Head = add
				q.Tail = add
			}
			q.Tasks[name] = add
		}
	}
}

// 后插队(当前新增任务添加到after任务之后)
func (q *QueueScheduler) AddAfter(name string, task func() error, after string) {
	if name != "" && task != nil {
		q.mutex.Lock()
		defer q.mutex.Unlock()
		if _, exist := q.Tasks[name]; !exist {
			var add = &QueueTask{name: name, fn: task}
			if target, ok := q.Tasks[after]; ok && target.next != nil {
				add.next = target.next
				add.prev = target
				target.next.prev = add
				target.next = add
			} else {
				target = q.Tail
				target.next = add
				add.prev = target
				q.Tail = add
			}
			q.Tasks[name] = add
		}
	}
}

// 前插队(当前新增任务添加到before任务之后)
func (q *QueueScheduler) AddBefore(name string, task func() error, before string) {
	if name != "" && task != nil {
		q.mutex.Lock()
		defer q.mutex.Unlock()
		if _, exist := q.Tasks[name]; !exist {
			var add = &QueueTask{name: name, fn: task}
			if target, ok := q.Tasks[before]; ok && target.prev != nil {
				add.prev = target.prev
				add.next = target
				target.prev.next = add
				target.prev = add
			} else {
				target = q.Head
				target.prev = add
				add.next = target
				q.Head = add
			}
			q.Tasks[name] = add
		}
	}
}

// 移除任务
func (q *QueueScheduler) Remove(name string) {
	if name != "" {
		q.mutex.Lock()
		defer q.mutex.Unlock()
		if task, ok := q.Tasks[name]; ok {
			if task.prev == nil {
				task.next.prev = nil
				q.Head = task.next
			} else if task.next == nil {
				task.prev.next = nil
				q.Tail = task.prev
			} else {
				task.prev.next = task.next
				task.next.prev = task.prev
			}
			delete(q.Tasks, name)
		}
	}
}

// 清除所有任务
func (q *QueueScheduler) Clear() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.Head = nil
	q.Tail = nil
	q.Tasks = make(map[string]*QueueTask)
}