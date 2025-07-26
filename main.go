package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// --- Структура данных для Задачи (Task) ---
type Task struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"` // e.g., "высокий", "средний", "низкий"
	DueDate     *time.Time `json:"dueDate"`  // Optional due date
	Tags        string     `json:"tags"`     // Comma-separated tags, e.g., "проект X, срочно"
	IsCompleted bool       `json:"isCompleted"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

var db *gorm.DB // Глобальная переменная для подключения к БД

// --- Инициализация базы данных ---
func initDB() {
	// Загружаем переменные окружения из .env файла (только для локальной разработки)
	// В Docker Compose они будут передаваться напрямую через environment
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, assuming environment variables are set.")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set.")
	}

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully.")

	// Автоматическая миграция схемы базы данных
	// GORM создаст таблицы, если их нет
	err = db.AutoMigrate(&Task{})
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
	}
	log.Println("Database migration completed.")
}

// --- Обработчики API для задач (CRUD) ---

// CreateTask - Создать новую задачу
func CreateTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Create(&task)
	c.JSON(http.StatusCreated, task)
}

// GetTasks - Получить список всех задач
func GetTasks(c *gin.Context) {
	var tasks []Task
	db.Find(&tasks)
	c.JSON(http.StatusOK, tasks)
}

// GetTaskByID - Получить задачу по ID
func GetTaskByID(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if result := db.First(&task, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, task)
}

// UpdateTask - Обновить существующую задачу
func UpdateTask(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if result := db.First(&task, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	db.Save(&task)
	c.JSON(http.StatusOK, task)
}

// DeleteTask - Удалить задачу
func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	var task Task
	if result := db.First(&task, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	db.Delete(&task)
	c.JSON(http.StatusNoContent, nil)
}

// --- Интеграция с ИИ-агентом (Заглушка) ---
// AIProcessQuery - Конечная точка для обработки запросов к ИИ-агенту
func AIProcessQuery(c *gin.Context) {
	var requestBody struct {
		Query string `json:"query" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query is required"})
		return
	}

	userQuery := requestBody.Query
	log.Printf("Received AI query: \"%s\"", userQuery)

	// --- Здесь будет ваша основная логика ИИ-агента ---
	// 1. Отправка запроса в LLM API (OpenAI, Google Gemini и т.д.)
	// 2. Интерпретация ответа LLM для получения критериев фильтрации
	// 3. Фильтрация задач из базы данных на основе полученных критериев

	var filteredTasks []Task
	// Просто пример фильтрации по ключевым словам для демонстрации
	if userQuery == "покажи срочные" {
		db.Where("priority = ?", "высокий").Find(&filteredTasks)
	} else if userQuery == "покажи завершенные" {
		db.Where("is_completed = ?", true).Find(&filteredTasks)
	} else if userQuery == "покажи незавершенные" {
		db.Where("is_completed = ?", false).Find(&filteredTasks)
	} else {
		db.Find(&filteredTasks)
		log.Println("AI could not provide specific filters, returning all tasks (or implement LLM clarification).")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       fmt.Sprintf("Processing AI query: '%s'", userQuery),
		"filteredTasks": filteredTasks,
		"note":          "AI logic is currently a placeholder. Implement LLM API calls and robust filtering here.",
	})
}

// --- Главная функция ---
func main() {
	initDB() // Инициализация базы данных при запуске приложения

	router := gin.Default()

	// Ping-маршрут (для проверки доступности сервера)
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Группировка маршрутов для API задач
	tasksGroup := router.Group("/tasks")
	{
		tasksGroup.POST("/", CreateTask)
		tasksGroup.GET("/", GetTasks)
		tasksGroup.GET("/:id", GetTaskByID)
		tasksGroup.PUT("/:id", UpdateTask)
		tasksGroup.DELETE("/:id", DeleteTask)
	}

	// Маршрут для ИИ-агента
	router.POST("/ai/query", AIProcessQuery)

	log.Fatal(router.Run(":8080")) // Запуск сервера на порту 8080
}
