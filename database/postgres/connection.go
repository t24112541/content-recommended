package postgres

import (
	"flag"
	"fmt"
	"log"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/fatih/color"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     string `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	Name     string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSL_MODE"`
}

func NewConnection() (conf DBConfig, err error) {
	err = env.Parse(&conf)

	return
}

func (conf *DBConfig) Connect() (db *gorm.DB, err error) {
	var once sync.Once
	once.Do(func() {
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			conf.Host,
			conf.Port,
			conf.User,
			conf.Password,
			conf.Name,
			conf.SSLMode,
		)

		db, err = gorm.Open(
			postgres.Open(dsn),
			&gorm.Config{},
		)
		if err != nil {
			log.Fatalf("Failed to connect to database: %s", color.RedString(err.Error()))
		}

		migrate := flag.Bool("migrate", false, "Automatically migrate the database")
		seed := flag.Bool("seed", false, "Automatically seeding data")

		flag.Parse()

		rDB := NewResourceDB(db)
		if *migrate {
			fmt.Print("Database migrate: ")

			if err := rDB.migrate(); err != nil {
				color.Red("failed")
				log.Fatalf("Failed to migrate database: %s", color.RedString(err.Error()))
			} else {
				color.Green("successfully")
			}
		}

		if *seed {
			fmt.Print("Sedding data: ")

			if err := rDB.seedData(); err != nil {
				color.Red("failed")
			} else {
				color.Green("successfully")
			}
		}
	})

	return db, err
}
