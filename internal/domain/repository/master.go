package repository

type MasterRepository interface {
	TaskDispatchLoop(interval int)
	CheckTaskStatusLoop(interval int)
	PullTaskResultLoop(interval int)
	CheckTaskIsOverLoop(interval int)
}
