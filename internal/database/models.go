package database

import "time"

// DataFormat representa um formato de livro disponível (ex: EPUB, PDF, MOBI).
type DataFormat struct {
	Format string `json:"format"`
	Size   int64  `json:"size"`
	Name   string `json:"name"`
}

// Book representa a entidade de livro agregada.
type Book struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Sort        string       `json:"sort"`
	Path        string       `json:"path"`
	HasCover    bool         `json:"has_cover"`
	PubDate     time.Time    `json:"pubdate"`
	SeriesIndex float64      `json:"series_index"`
	Authors     []string     `json:"authors"`
	Series      string       `json:"series,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
	Formats     []DataFormat `json:"formats,omitempty"`
	Publishers  []string     `json:"publishers,omitempty"`
	Rating      int          `json:"rating,omitempty"`
	Languages   []string     `json:"languages,omitempty"`
	ReadStatus  bool         `json:"read_status"`
	IsArchived  bool         `json:"is_archived"`
}

// BookDetails inclui metadados completos como a sinopse e identificadores externos.
type BookDetails struct {
	Book
	Comment     string            `json:"comment"`
	Identifiers map[string]string `json:"identifiers"`
}

// PageResult estrutura de resposta paginada forçada (Zero Heap Spillover).
type PageResult[T any] struct {
	Items      []T   `json:"items"`
	TotalItems int64 `json:"total_items"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// BookFilter define critérios de paginação e pesquisa completos do Calibre-Web.
type BookFilter struct {
	Page        int
	PageSize    int
	Search      string
	AuthorID    int64
	TagID       int64
	SeriesID    int64
	PublisherID int64
	RatingID    int64
	LanguageID  int64
	Format      string
	Sort        string
	ReadStatus  int // 0 = all, 1 = read, 2 = unread
	UserID      int64
	BookIDs     []int64
}

// EntityCount representa uma entidade (autor, série, categoria, formato, etc.) com contagem de livros.
type EntityCount struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// User representa um utilizador do sistema Calibre-Web.
type User struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	Email           string `json:"email"`
	KindleMail      string `json:"kindle_mail"`
	Role            int    `json:"role"`
	IsAdmin         bool   `json:"is_admin"`
	Downloads       int    `json:"downloads"`
	Locale          string `json:"locale"`
	DefaultLanguage string `json:"default_language"`
	PasswordHash    string `json:"-"`
}

// Shelf representa uma estante personalizada do utilizador.
type Shelf struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	IsPublic  bool   `json:"is_public"`
	BookCount int    `json:"book_count"`
}

// MagicShelf representa uma estante dinâmica baseada em regras inteligentes.
type MagicShelf struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Rules       string `json:"rules"`
	IsPublic    bool   `json:"is_public"`
	UserID      int64  `json:"user_id"`
	BookCount   int    `json:"book_count"`
}

// TaskItem representa uma tarefa assíncrona em execução ou finalizada.
type TaskItem struct {
	ID          string    `json:"id"`
	User        string    `json:"user"`
	TaskMessage string    `json:"taskMessage"`
	Status      string    `json:"status"`
	Progress    string    `json:"progress"`
	RunTime     string    `json:"runtime"`
	StartTime   time.Time `json:"starttime"`
}

// BookEditForm contém os metadados a atualizar na edição de um livro.
type BookEditForm struct {
	Title       string
	TitleSort   string
	Authors     []string
	AuthorSort  string
	Series      string
	SeriesIndex float64
	Tags        []string
	Publishers  []string
	Languages   []string
	Rating      int
	Comment     string
	Identifiers map[string]string
}

// DuplicateGroup representa um grupo de livros potencialmente duplicados com percentagem de confiança.
type DuplicateGroup struct {
	Hash       string        `json:"hash"`
	Title      string        `json:"title"`
	Confidence int           `json:"confidence"`
	Books      []BookDetails `json:"books"`
}

// KOSyncProgress armazena o progresso de leitura sincronizado com o KOReader.
type KOSyncProgress struct {
	UserID     int64   `json:"user_id"`
	Document   string  `json:"document"`
	Progress   string  `json:"progress"`
	Percentage float64 `json:"percentage"`
	Device     string  `json:"device"`
	DeviceID   string  `json:"device_id"`
	Timestamp  int64   `json:"timestamp"`
}

// SMTPSettings armazena os parâmetros de envio de email extraídos da tabela settings.
type SMTPSettings struct {
	Server   string
	Port     int
	UseSSL   int
	Login    string
	Password string
	From     string
}

