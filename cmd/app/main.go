package main

import (
	"github.com/azmiagr/lumbera-hackathon/internal/handler/rest"
	"github.com/azmiagr/lumbera-hackathon/internal/repository"
	"github.com/azmiagr/lumbera-hackathon/internal/service"
	"github.com/azmiagr/lumbera-hackathon/pkg/bcrypt"
	"github.com/azmiagr/lumbera-hackathon/pkg/config"
	"github.com/azmiagr/lumbera-hackathon/pkg/database/mariadb"
	"github.com/azmiagr/lumbera-hackathon/pkg/jwt"
	"github.com/azmiagr/lumbera-hackathon/pkg/middleware"
	"log"
)

func main() {
	config.LoadEnvironment()

	db, err := mariadb.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	err = mariadb.Migrate(db)
	if err != nil {
		log.Fatal(err)
	}

	repo := repository.NewRepository(db)
	bcrypt := bcrypt.Init()
	jwt := jwt.Init()
	svc := service.NewService(repo, bcrypt, jwt)

	middleware := middleware.Init(svc, jwt)
	r := rest.NewRest(svc, middleware)
	r.MountEndpoint()

	r.Run()
}
