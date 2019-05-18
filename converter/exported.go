package converter

import (
	// stdlib
	"encoding/json"
	"flag"
	"log"
	"sync"
	"time"

	// local
	"github.com/pztrn/ffmpeger/nats"
)

var (
	// ffmpeg path.
	ffmpegPath string

	// Tasks queue.
	tasks      []*Task
	tasksMutex sync.Mutex

	// Currently running tasks.
	// Reason why this isn't from atomic package is because atomic's
	// integers (as well as other things) doesn't neccessarily changed
	// when Add* functions called but we need to make sure that our
	// running count is precise.
	// Mutex is here because value will be decremented/incremented from
	// worker goroutine and read from control goroutine.
	currentlyRunning      int
	currentlyRunningMutex sync.Mutex

	// Maximum tasks that should be executed concurrently.
	// No mutex here because it will be accessed from only one place
	// after initialization.
	maximumConcurrentTasks int

	// Indicates that we should shutdown working goroutine.
	shouldShutdown      bool
	shouldShutdownMutex sync.Mutex

	// Indicates that goroutine was successfully shutdown.
	shuttedDown chan bool
)

// AddTask adds task to processing queue.
func AddTask(task *Task) {
	tasksMutex.Lock()
	tasks = append(tasks, task)
	tasksMutex.Unlock()
}

// Initialize initializes package.
func Initialize() {
	log.Println("Initializing converter...")

	tasks = make([]*Task, 0, 64)
	shuttedDown = make(chan bool, 1)

	flag.IntVar(&maximumConcurrentTasks, "maxconcurrency", 1, "Maximum conversion tasks that should be run concurrently")

	handler := &nats.Handler{
		Name: "converter",
		Func: natsMessageHandler,
	}
	nats.AddHandler(handler)
}

func natsMessageHandler(data []byte) {
	t := &Task{}
	json.Unmarshal(data, t)
	log.Printf("Received task: %+v\n", t)

	tasksMutex.Lock()
	tasks = append(tasks, t)
	tasksMutex.Unlock()
}

// Shutdown sets shutdown flag and waits until shuttedDown channel will
// get any message means that shutdown was completed.
func Shutdown() {
	log.Println("Starting converter shutdown...")
	shouldShutdownMutex.Lock()
	shouldShutdown = true
	shouldShutdownMutex.Unlock()

	<-shuttedDown
	log.Println("Converter shutted down")
}

// Start starts working goroutine.
func Start() {
	log.Println("Starting converter controlling goroutine...")
	log.Println("Maximum simultaneous tasks to run:", maximumConcurrentTasks)
	findffmpeg()

	go startReally()
}

// Real start for working goroutine.
func startReally() {
	tick := time.NewTicker(time.Second * 1)
	for range tick.C {
		// Check for shutdown.
		// Boolean values aren't goroutine-safe that's why we create local
		// copy of package variable.
		shouldShutdownMutex.Lock()
		weHaveToShutdown := shouldShutdown
		shouldShutdownMutex.Unlock()

		if weHaveToShutdown {
			log.Println("Stopping tasks distribution...")
			break
		}

		// Check for tasks available and currently running counts.
		currentlyRunningMutex.Lock()
		curRunning := currentlyRunning
		currentlyRunningMutex.Unlock()

		// Skip iteration if we have maximum tasks launched.
		if curRunning >= maximumConcurrentTasks {
			continue
		}

		// Check if we have tasks at all.
		tasksMutex.Lock()
		tasksCount := len(tasks)
		tasksMutex.Unlock()
		if tasksCount == 0 {
			log.Println("No tasks to launch")
			continue
		}

		// If we're here - we should launch a task! Lets get them.
		tasksToRunCount := maximumConcurrentTasks - curRunning
		tasksToRun := make([]*Task, 0, tasksToRunCount)
		tasksMutex.Lock()
		// To ensure that our tasks queue will be clean we will copy
		// queue, clear it and re-add if queue items still be there.
		tasksQueue := make([]*Task, 0, tasksCount)
		tasksQueue = append(tasksQueue, tasks...)
		tasks = make([]*Task, 0, 64)
		tasksMutex.Unlock()

		// Get tasks list to launch.
		for taskID, task := range tasksQueue {
			if taskID == tasksToRunCount {
				break
			}
			tasksToRun = append(tasksToRun, task)
		}
		// Remove tasks that will be launched now.
		tasksQueue = tasksQueue[tasksToRunCount:]
		log.Println("Tasks count that will be returned to main queue:", len(tasksQueue))
		// Re-add remaining tasks to queue.
		// Note: if another task was added to queue while we compose
		// our tasks list to launch - it will be executed BEFORE remaining
		// tasks.
		tasksMutex.Lock()
		tasks = append(tasks, tasksQueue...)
		tasksMutex.Unlock()

		log.Println("Got", len(tasksToRun), "tasks to run")

		// Launch tasks.
		for _, task := range tasksToRun {
			go task.Convert()
		}
	}

	// Waiting until all child goroutines will also shut down.
	log.Println("Waiting for all child goroutines to stop...")
	shutdownTicker := time.NewTicker(time.Millisecond * 500)
	for range shutdownTicker.C {
		currentlyRunningMutex.Lock()
		curRunning := currentlyRunning
		currentlyRunningMutex.Unlock()

		log.Println("Currently running converter goroutines:", curRunning)
		if curRunning == 0 {
			break
		}
	}

	shuttedDown <- true
}
