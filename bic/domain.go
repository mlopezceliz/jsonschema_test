package bic

type StructProducerConfig struct {
	ID              string                 `json:"id"`
	ProducerName    string                 `json:"producer_name"`
	Entity          string                 `json:"entity"`
	Status          string                 `json:"status"`
	AllowGet        bool                   `json:"allow_get"`
	SkipValidation  bool                   `json:"skip_validation"`
	ProductionID    *string                `json:"production_id"`
	AllowedMetrics  map[string]interface{} `json:"allowed_metrics"`
	FlowConfig      FlowConfig             `json:"flow_config"`
	MandatoryFields *[]string              `json:"mandatory_fields"`
	CreatedAt       string                 `json:"created_at"`
	CreatedBy       string                 `json:"created_by"`
	UpdatedAt       *string                `json:"updated_at"`
	UpdatedBy       *string                `json:"updated_by"`
}

type FlowConfig struct {
	BigQueueTopic      string    `json:"big_queue_topic"`
	Decorations        *[]string `json:"decorations"`
	OneTimeDecorations *[]string `json:"one_time_decorations"`
	Outputs            Outputs   `json:"outputs"`
}

type StructPayload struct {
	Entity        string                 `json:"entity"`
	ID            string                 `json:"id"`
	Metrics       map[string]interface{} `json:"metrics"`
	ProducerToken string
}

type Outputs struct {
	IndexNames     *[]string   `json:"index_names"`
	BiCoreInbounds *[]string   `json:"bicore_inbounds"`
	KvsDbNames     *[]string   `json:"kvs_db_names"`
	S3BucketNames  *[]string   `json:"s3_bucket_names"`
	S3Exports      *[]S3Export `json:"s3_exports"`
	KvsDsNames     *[]string   `json:"kvs_ds_names"`
}

type S3Export struct {
	KinesisCode  string    `json:"code"`
	ExportFields *[]string `json:"export_fields"`
	Format       string    `json:"format"`
}
