package util

const (
	TaskPathPrefix      = "/api/v1/task/"
	AlgorithmPathPrefix = "/api/v1/algorithms/"
	TokenPathPrefix     = "/api/v1/auth/"
	AlgorithmListPath   = AlgorithmPathPrefix
	SyncPreviewPath     = TaskPathPrefix + "preview/"
	RealtimeInitPath    = TaskPathPrefix + "init/"
	RealtimeRunPath     = TaskPathPrefix + "run/"
	RealtimeResultPath  = TaskPathPrefix + "result/"
	GenerateTokenPath   = TokenPathPrefix

	SyncPreviewType    = "syncPreview"
	RealtimeRunType    = "realtimeCheck"
	AlgorithmListType  = "algorithmList"
	RealtimeInitType   = "generateTaskId"
	RealtimeResultType = "realtimeResult"
	GenerateTokenType  = "generateToken"

	MetricsType    = "metrics"
	LabelNamesType = "labelNames"
	SeriesType     = "series"

	ProjectType = "grafana"
)
