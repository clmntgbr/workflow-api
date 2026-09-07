package steprun

import "time"

func ElapsedMS(startedAt, finishedAt *time.Time) int64 {
	if startedAt == nil || finishedAt == nil {
		return 0
	}
	d := finishedAt.Sub(*startedAt)
	if d <= 0 {
		return 0
	}
	return d.Milliseconds()
}

func (v StepRunView) ExecutionElapsedMS() int64 {
	return ElapsedMS(v.StartedAt, v.FinishedAt)
}

func SumExecutionElapsedMS(views []StepRunView) int64 {
	var total int64
	for _, view := range views {
		total += view.ExecutionElapsedMS()
	}
	return total
}
