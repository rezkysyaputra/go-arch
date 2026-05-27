package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"unicode"
)

type feature struct {
	Name       string
	Plural     string
	Pascal     string
	PascalMany string
	Table      string
	Migration  string
}

type fileTemplate struct {
	path string
	body string
}

func main() {
	name := flag.String("name", "", "feature name in singular form, for example product")
	plural := flag.String("plural", "", "optional plural form, for example products")
	flag.Parse()

	if *name == "" {
		exit("feature name is required, example: go run ./tools/featuregen -name product")
	}

	feature, err := newFeature(*name, *plural)
	if err != nil {
		exit(err.Error())
	}

	files := []fileTemplate{
		{path: fmt.Sprintf("internal/domain/%s.go", feature.Name), body: domainTemplate},
		{path: fmt.Sprintf("internal/usecase/%s_usecase.go", feature.Name), body: usecaseTemplate},
		{path: fmt.Sprintf("internal/repository/%s_repository.go", feature.Name), body: repositoryTemplate},
		{path: fmt.Sprintf("internal/delivery/http/%s_handler.go", feature.Name), body: handlerTemplate},
		{path: fmt.Sprintf("migrations/%s_create_%s_table.up.sql", feature.Migration, feature.Table), body: migrationUpTemplate},
		{path: fmt.Sprintf("migrations/%s_create_%s_table.down.sql", feature.Migration, feature.Table), body: migrationDownTemplate},
	}

	for _, file := range files {
		if err := writeTemplate(file.path, file.body, feature); err != nil {
			exit(err.Error())
		}
		fmt.Printf("created %s\n", file.path)
	}

	fmt.Println()
	fmt.Println("Next manual wiring steps:")
	fmt.Printf("- Add repository.NewGorm%sRepository(db) and usecase.New%sUsecase(...) in cmd/api/main.go\n", feature.Pascal, feature.Pascal)
	fmt.Printf("- Pass the %s usecase into internal/delivery/http.NewRouter\n", feature.Name)
	fmt.Printf("- Call Register%sRoutes(router, %sUsecase) in internal/delivery/http/router.go\n", feature.Pascal, feature.Name)
	fmt.Println("- Run gofmt, go test ./..., and go run ./cmd/migrate up")
}

func newFeature(name, plural string) (feature, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	plural = strings.TrimSpace(strings.ToLower(plural))

	validName := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !validName.MatchString(name) {
		return feature{}, fmt.Errorf("feature name must use snake_case, for example product or order_item")
	}

	if plural == "" {
		plural = name + "s"
	}
	if !validName.MatchString(plural) {
		return feature{}, fmt.Errorf("plural name must use snake_case")
	}

	migration, err := nextMigrationVersion("migrations")
	if err != nil {
		return feature{}, err
	}

	return feature{
		Name:       name,
		Plural:     plural,
		Pascal:     pascal(name),
		PascalMany: pascal(plural),
		Table:      plural,
		Migration:  migration,
	}, nil
}

func nextMigrationVersion(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("read migrations directory: %w", err)
	}

	maxVersion := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if len(name) < 6 {
			continue
		}

		version, err := strconv.Atoi(name[:6])
		if err == nil && version > maxVersion {
			maxVersion = version
		}
	}

	return fmt.Sprintf("%06d", maxVersion+1), nil
}

func writeTemplate(path, body string, data feature) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create directory for %s: %w", path, err)
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "close %s: %v\n", path, closeErr)
		}
	}()

	tmpl, err := template.New(filepath.Base(path)).Parse(body)
	if err != nil {
		return fmt.Errorf("parse template for %s: %w", path, err)
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("execute template for %s: %w", path, err)
	}

	return nil
}

func pascal(value string) string {
	parts := strings.Split(value, "_")
	for index, part := range parts {
		if part == "" {
			continue
		}

		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		parts[index] = string(runes)
	}

	return strings.Join(parts, "")
}

func exit(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

const domainTemplate = `package domain

import (
	"context"
	"errors"
	"time"
)

var Err{{ .Pascal }}NotFound = errors.New("{{ .Name }} not found")

type {{ .Pascal }} struct {
	ID        uint
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Create{{ .Pascal }}Input struct {
	Name string
}

type {{ .Pascal }}Repository interface {
	Create(ctx context.Context, {{ .Name }} *{{ .Pascal }}) error
	FindByID(ctx context.Context, id uint) (*{{ .Pascal }}, error)
}

type {{ .Pascal }}Usecase interface {
	Create(ctx context.Context, input Create{{ .Pascal }}Input) (*{{ .Pascal }}, error)
	GetByID(ctx context.Context, id uint) (*{{ .Pascal }}, error)
}
`

const usecaseTemplate = `package usecase

import (
	"context"
	"strings"

	"go-arch/internal/domain"
)

type {{ .Pascal }}Usecase struct {
	{{ .Plural }} domain.{{ .Pascal }}Repository
}

func New{{ .Pascal }}Usecase({{ .Plural }} domain.{{ .Pascal }}Repository) *{{ .Pascal }}Usecase {
	return &{{ .Pascal }}Usecase{
		{{ .Plural }}: {{ .Plural }},
	}
}

func (uc *{{ .Pascal }}Usecase) Create(ctx context.Context, input domain.Create{{ .Pascal }}Input) (*domain.{{ .Pascal }}, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, domain.ErrInvalidInput
	}

	{{ .Name }} := &domain.{{ .Pascal }}{
		Name: input.Name,
	}

	if err := uc.{{ .Plural }}.Create(ctx, {{ .Name }}); err != nil {
		return nil, err
	}

	return {{ .Name }}, nil
}

func (uc *{{ .Pascal }}Usecase) GetByID(ctx context.Context, id uint) (*domain.{{ .Pascal }}, error) {
	if id == 0 {
		return nil, domain.ErrInvalidInput
	}

	return uc.{{ .Plural }}.FindByID(ctx, id)
}
`

const repositoryTemplate = `package repository

import (
	"context"
	"errors"
	"time"

	"go-arch/internal/domain"

	"gorm.io/gorm"
)

type {{ .Name }}Model struct {
	ID        uint      ` + "`gorm:\"primaryKey\"`" + `
	Name      string    ` + "`gorm:\"size:255;not null\"`" + `
	CreatedAt time.Time ` + "`gorm:\"not null\"`" + `
	UpdatedAt time.Time ` + "`gorm:\"not null\"`" + `
}

func ({{ .Name }}Model) TableName() string {
	return "{{ .Table }}"
}

type Gorm{{ .Pascal }}Repository struct {
	db *gorm.DB
}

func NewGorm{{ .Pascal }}Repository(db *gorm.DB) *Gorm{{ .Pascal }}Repository {
	return &Gorm{{ .Pascal }}Repository{db: db}
}

func (r *Gorm{{ .Pascal }}Repository) Create(ctx context.Context, {{ .Name }} *domain.{{ .Pascal }}) error {
	model := {{ .Name }}ModelFromDomain({{ .Name }})
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	*{{ .Name }} = *model.toDomain()
	return nil
}

func (r *Gorm{{ .Pascal }}Repository) FindByID(ctx context.Context, id uint) (*domain.{{ .Pascal }}, error) {
	var {{ .Name }} {{ .Name }}Model
	err := r.db.WithContext(ctx).First(&{{ .Name }}, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.Err{{ .Pascal }}NotFound
	}
	if err != nil {
		return nil, err
	}

	return {{ .Name }}.toDomain(), nil
}

func {{ .Name }}ModelFromDomain({{ .Name }} *domain.{{ .Pascal }}) {{ .Name }}Model {
	return {{ .Name }}Model{
		ID:        {{ .Name }}.ID,
		Name:      {{ .Name }}.Name,
		CreatedAt: {{ .Name }}.CreatedAt,
		UpdatedAt: {{ .Name }}.UpdatedAt,
	}
}

func (m {{ .Name }}Model) toDomain() *domain.{{ .Pascal }} {
	return &domain.{{ .Pascal }}{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}
`

const handlerTemplate = `package http

import (
	"errors"
	nethttp "net/http"
	"strconv"
	"time"

	"go-arch/internal/domain"

	"github.com/gin-gonic/gin"
)

type {{ .Pascal }}Handler struct {
	{{ .Plural }} domain.{{ .Pascal }}Usecase
}

func New{{ .Pascal }}Handler({{ .Plural }} domain.{{ .Pascal }}Usecase) *{{ .Pascal }}Handler {
	return &{{ .Pascal }}Handler{ {{ .Plural }}: {{ .Plural }} }
}

func Register{{ .Pascal }}Routes(router *gin.Engine, {{ .Plural }} domain.{{ .Pascal }}Usecase) {
	handler := New{{ .Pascal }}Handler({{ .Plural }})

	group := router.Group("/{{ .Plural }}")
	{
		group.POST("", handler.Create)
		group.GET("/:id", handler.GetByID)
	}
}

func (h *{{ .Pascal }}Handler) Create(ctx *gin.Context) {
	var request create{{ .Pascal }}Request
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondValidationError(ctx, err)
		return
	}

	{{ .Name }}, err := h.{{ .Plural }}.Create(ctx.Request.Context(), domain.Create{{ .Pascal }}Input{
		Name: request.Name,
	})
	if err != nil {
		respond{{ .Pascal }}Error(ctx, err)
		return
	}

	respondSuccess(ctx, nethttp.StatusCreated, "{{ .Name }} created successfully", new{{ .Pascal }}Response({{ .Name }}))
}

func (h *{{ .Pascal }}Handler) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		respondFailure(ctx, nethttp.StatusBadRequest, "invalid request", []errorDetail{
			{Field: "id", Code: "invalid_id", Message: "id must be a positive integer"},
		})
		return
	}

	{{ .Name }}, err := h.{{ .Plural }}.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		respond{{ .Pascal }}Error(ctx, err)
		return
	}

	respondSuccess(ctx, nethttp.StatusOK, "{{ .Name }} retrieved successfully", new{{ .Pascal }}Response({{ .Name }}))
}

func respond{{ .Pascal }}Error(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		respondFailure(ctx, nethttp.StatusBadRequest, "invalid request", []errorDetail{
			{Code: "invalid_input", Message: err.Error()},
		})
	case errors.Is(err, domain.Err{{ .Pascal }}NotFound):
		respondFailure(ctx, nethttp.StatusNotFound, "resource not found", []errorDetail{
			{Code: "{{ .Name }}_not_found", Message: err.Error()},
		})
	default:
		respondFailure(ctx, nethttp.StatusInternalServerError, "internal server error", []errorDetail{
			{Code: "internal_error", Message: "unexpected server error"},
		})
	}
}

type create{{ .Pascal }}Request struct {
	Name string ` + "`json:\"name\" binding:\"required,min=2,max=255\"`" + `
}

type {{ .Name }}Response struct {
	ID        uint   ` + "`json:\"id\"`" + `
	Name      string ` + "`json:\"name\"`" + `
	CreatedAt string ` + "`json:\"created_at\"`" + `
	UpdatedAt string ` + "`json:\"updated_at\"`" + `
}

func new{{ .Pascal }}Response({{ .Name }} *domain.{{ .Pascal }}) {{ .Name }}Response {
	return {{ .Name }}Response{
		ID:        {{ .Name }}.ID,
		Name:      {{ .Name }}.Name,
		CreatedAt: {{ .Name }}.CreatedAt.Format(time.RFC3339),
		UpdatedAt: {{ .Name }}.UpdatedAt.Format(time.RFC3339),
	}
}
`

const migrationUpTemplate = `CREATE TABLE IF NOT EXISTS {{ .Table }} (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
`

const migrationDownTemplate = `DROP TABLE IF EXISTS {{ .Table }};
`
