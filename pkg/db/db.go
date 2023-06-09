package db

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	utils "github.com/infor-design/selfservice/pkg/utils"
	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

type Connection struct {
	*gorm.DB
}

func NewConfig() *Config {
	return &Config{
		Host:     utils.GetEnv("DB_HOST", "postgres"),
		Port:     utils.GetEnv("DB_PORT", "5432"),
		User:     utils.GetEnv("DB_USER", "postgres"),
		Password: utils.GetEnv("DB_PASS", "postgres"),
		Database: utils.GetEnv("DB_NAME", "postgres"),
	}
}

func NewDb(cfg *Config) *Connection {
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Password == "" || cfg.Database == "" {
		panic(errors.Errorf("All fields must be set (%s)", spew.Sdump(cfg)))
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.Port,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		panic(err)
	} else {
		fmt.Println("Successfully connected to postgres")
	}

	return &Connection{db}
}

func (c *Connection) Close() (err error) {
	if c == nil {
		return
	}

	if err = c.Close(); err != nil {
		panic(err)
	}

	return
}

func (c *Connection) InitialMigration() {
	c.AutoMigrate(&Application{})
	c.AutoMigrate(&Job{})
	c.AutoMigrate(&Repo{})

	if !c.Migrator().HasConstraint(&Application{}, "Jobs") {
		c.Migrator().CreateConstraint(&Application{}, "Jobs")
	}
}
