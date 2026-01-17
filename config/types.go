package config

type BlockType string

const (
	BlockTypeValue  BlockType = "value"  // ValueBlock
	BlockTypeHeader BlockType = "header" // HeaderBlock
	BlockTypeMatrix BlockType = "matrix" // MatrixBlock
)

type Direction string

const (
	DirectionVertical   Direction = "vertical"
	DirectionHorizontal Direction = "horizontal"
)

type CellRange struct {
	Ref string `json:"ref" yaml:"ref"` // e.g. "A1:G33"
}

type DataSourceConfig struct {
	Name   string `json:"name"   yaml:"name"`
	Driver string `json:"driver" yaml:"driver"` // "mysql", "sqlserver", "postgres"
	DSN    string `json:"dsn"    yaml:"dsn"`    // 连接串
}

type LabelConfig struct {
	Name   string `json:"name"   yaml:"name"`   // label name
	Column string `json:"column" yaml:"column"` // actual column name in db
	Type   string `json:"type,omitempty" yaml:"type,omitempty"`
}

type DataViewConfig struct {
	Id         string        `json:"id"         yaml:"id"`
	Name       string        `json:"name"       yaml:"name"`
	DataSource string        `json:"dataSource" yaml:"dataSource"`
	Sql        string        `json:"sql,omitempty" yaml:"sql,omitempty"`
	Table      string        `json:"table,omitempty" yaml:"table,omitempty"`
	Labels     []LabelConfig `json:"labels" yaml:"labels"`
}

type BlockConfig struct {
	Name         string     `json:"name"          yaml:"name"`
	Type         BlockType  `json:"type"          yaml:"type"`  // value / header / matrix
	Range        CellRange  `json:"range"         yaml:"range"` // 块整体范围
	LabelRange   *CellRange `json:"labelRange,omitempty" yaml:"labelRange,omitempty"`
	DataViewName string     `json:"dataView,omitempty" yaml:"dataView,omitempty"` // 绑定的数据源视图名

	// MatrixBlock / HeaderBlock
	Direction     Direction `json:"direction,omitempty" yaml:"direction,omitempty"`
	RowLimit      int       `json:"rowLimit,omitempty" yaml:"rowLimit,omitempty"`
	InsertAfter   bool      `json:"insertAfter,omitempty" yaml:"insertAfter,omitempty"`
	LabelVariable string    `json:"labelVariable,omitempty" yaml:"labelVariable,omitempty"`

	// Template ValueBlock of MatrixBlock
	Template bool `json:"template,omitempty" yaml:"template,omitempty"`

	// Nested
	SubBlocks []BlockConfig `json:"subBlocks,omitempty" yaml:"subBlocks,omitempty"`
}

type SheetConfig struct {
	Name                string        `json:"name"         yaml:"name"`
	Dynamic             bool          `json:"dynamic"      yaml:"dynamic"`
	ParamLabel          string        `json:"paramLabel,omitempty" yaml:"paramLabel,omitempty"`
	DataViewName        string        `json:"dataView,omitempty" yaml:"dataView,omitempty"`
	VerticalArrangement bool          `json:"verticalArrangement" yaml:"verticalArrangement"`
	AllowOverlap        bool          `json:"allowOverlap" yaml:"allowOverlap"`
	Blocks              []BlockConfig `json:"blocks"       yaml:"blocks"`
}

type WorkbookConfig struct {
	Id          string            `json:"id"           yaml:"id"`
	Name        string            `json:"name"         yaml:"name"`
	Template    string            `json:"template"     yaml:"template"`
	OutputDir   string            `json:"outputDir"    yaml:"outputDir"`
	ArchiveRule string            `json:"archiveRule,omitempty" yaml:"archiveRule,omitempty"`
	Parameters  map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Sheets      []SheetConfig     `json:"sheets"       yaml:"sheets"`
}
