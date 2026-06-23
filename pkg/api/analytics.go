package api

type AnalyticsData struct {
	InvocationCount   int64
	InvocationBuckets []float64

	AvgWallTimeMs    float64
	SumWallTimeMs    int64
	WallTimeBuckets  []float64

	FailRatio   float64
	FailBuckets []float64

	TimeRange string
}
