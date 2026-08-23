package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"weblog-app/handler"
	customMiddleware "weblog-app/middleware"
	"weblog-app/repo"
	"weblog-app/service"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func getDBConnStr() string {
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		return dbURL
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "localhost"
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "postgres"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "password"
	}
	dbname := os.Getenv("DB_NAME")
	if dbname == "" {
		dbname = "weblogdb"
	}

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}

func initSchema(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(100) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS posts (
		id SERIAL PRIMARY KEY,
		author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		title VARCHAR(255) NOT NULL,
		content TEXT NOT NULL,
		image_url VARCHAR(255),
		is_private BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS post_shares (
		post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		PRIMARY KEY (post_id, user_id)
	);

	CREATE TABLE IF NOT EXISTS comments (
		id SERIAL PRIMARY KEY,
		post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
		author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		content TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);`
	_, err := db.Exec(schema)
	return err
}

func main() {
	_ = os.MkdirAll("uploads", os.ModePerm)

	db, err := sql.Open("postgres", getDBConnStr())
	if err != nil {
		log.Fatalf("Database connection error: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Database ping error: %v", err)
	}

	if err := initSchema(db); err != nil {
		log.Fatalf("Schema init error: %v", err)
	}

	userRepo := repo.NewUserRepo(db)
	postRepo := repo.NewPostRepo(db)
	commentRepo := repo.NewCommentRepo(db)

	authService := service.NewAuthService(userRepo)
	postService := service.NewPostService(postRepo, commentRepo, userRepo)

	authHandler := handler.NewAuthHandler(authService)
	postHandler := handler.NewPostHandler(postService, userRepo)

	e := echo.New()

	renderer := &TemplateRenderer{
		templates: template.Must(template.ParseGlob("templates/*.html")),
	}
	e.Renderer = renderer

	e.Static("/uploads", "uploads")

	// Public Routes
	e.GET("/login", authHandler.GetLogin)
	e.POST("/login", authHandler.PostLogin)
	e.GET("/signup", authHandler.GetSignup)
	e.POST("/signup", authHandler.PostSignup)

	// Protected Routes Group
	protected := e.Group("")
	protected.Use(customMiddleware.AuthMiddleware(authService))

	protected.GET("/", postHandler.GetFeed)
	protected.GET("/posts/new", postHandler.GetNewPostPage)
	protected.POST("/posts", postHandler.PostCreate)
	protected.GET("/weblog/:id", postHandler.GetPostDetail)
	protected.POST("/posts/:id/delete", postHandler.PostDelete)
	protected.POST("/weblog/:id/comments", postHandler.PostComment)
	protected.POST("/logout", authHandler.PostLogout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	e.Logger.Fatal(e.Start(":" + port))
}