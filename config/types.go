package config

type BlockType string

const (
	BlockTypeTag    BlockType = "T" // TagBlock
	BlockTypeAxis   BlockType = "A" // AxisBlock
	BlockTypeExpand BlockType = "E" // ExpandableBlock
)

type Direction string

const (
	DirectionVertical   Direction = "vertical"
	DirectionHorizontal Direction = "horizontal"
)

// CellRange： cell range config
type CellRange struct {
	Ref string `json:"ref" yaml:"ref"` // e.g. "A1:G33"
}

// DataSourceConfig：datasource config
type DataSourceConfig struct {
	Name   string `json:"name"   yaml:"name"`
	Driver string `json:"driver" yaml:"driver"` // "mysql", "sqlserver", "postgres"
	DSN    string `json:"dsn"    yaml:"dsn"`    // 连接串
}

// TagConfig：virtual view里的“标签字段”配置
type TagConfig struct {
	Name   string `json:"name"   yaml:"name"`   // tag name
	Column string `json:"column" yaml:"column"` // actual column name in db
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
}

// VirtualViewConfig：virtual view config
type VirtualViewConfig struct {
	Id         string      `json:"id"         yaml:"id"`
	Name       string      `json:"name"       yaml:"name"`
	DataSource string      `json:"dataSource" yaml:"dataSource"`
	Sql        string      `json:"sql,omitempty" yaml:"sql,omitempty"`
	Table      string      `json:"table,omitempty" yaml:"table,omitempty"`
	Tags       []TagConfig `json:"tags"       yaml:"tags"`
}

// BlockConfig：config for a block
type BlockConfig struct {
	Name      string     `json:"name"          yaml:"name"`
	Type      BlockType  `json:"type"          yaml:"type"`  // T / A / E
	Range     CellRange  `json:"range"         yaml:"range"` // 块整体范围
	TagRange  *CellRange `json:"tagRange,omitempty" yaml:"tagRange,omitempty"`
	VViewName string     `json:"vview,omitempty" yaml:"vview,omitempty"` // 绑定的数据源视图名

	// Expandable / Axis
	Direction   Direction `json:"direction,omitempty" yaml:"direction,omitempty"`
	RowLimit    int       `json:"rowLimit,omitempty" yaml:"rowLimit,omitempty"`
	InsertAfter bool      `json:"insertAfter,omitempty" yaml:"insertAfter,omitempty"`
	TagVariable string    `json:"tagVariable,omitempty" yaml:"tagVariable,omitempty"`

	// Template TagBlock of ExpandableBlock
	Template bool `json:"template,omitempty" yaml:"template,omitempty"`

	// Nested
	SubBlocks []BlockConfig `json:"subBlocks,omitempty" yaml:"subBlocks,omitempty"`
}

// SheetConfig：sheet range config
type SheetConfig struct {
	Name                string        `json:"name"         yaml:"name"`
	Dynamic             bool          `json:"dynamic"      yaml:"dynamic"`
	ParamTag            string        `json:"paramTag,omitempty" yaml:"paramTag,omitempty"`
	VViewName           string        `json:"vview,omitempty" yaml:"vview,omitempty"`
	VerticalArrangement bool          `json:"verticalArrangement" yaml:"verticalArrangement"`
	AllowOverlap        bool          `json:"allowOverlap" yaml:"allowOverlap"`
	Blocks              []BlockConfig `json:"blocks"       yaml:"blocks"`
}

// WorkbookConfig：workbook range config
type WorkbookConfig struct {
	Id          string            `json:"id"           yaml:"id"`
	Name        string            `json:"name"         yaml:"name"`
	Template    string            `json:"template"     yaml:"template"`
	OutputDir   string            `json:"outputDir"    yaml:"outputDir"`
	ArchiveRule string            `json:"archiveRule,omitempty" yaml:"archiveRule,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Sheets      []SheetConfig     `json:"sheets"       yaml:"sheets"`
}
