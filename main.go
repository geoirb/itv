package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/GeoIrb/itv/app"
	"github.com/GeoIrb/itv/controllers"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	conn := app.Init()
	defer conn.Stop()

	e := echo.New()

	e.Use(middleware.Recover())
	// e.Use(middleware.Logger())
	e.Use(middleware.Gzip())
	e.Use(middleware.Secure())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.OPTIONS, echo.GET, echo.PUT, echo.POST, echo.DELETE},
	}))
	e.Use(middleware.Static("react/build"))

	taskc := controllers.NewTaskController(conn)
	e.POST("/request", taskc.FetchTask)
	e.POST("/request/chan", taskc.FetchTaskChan)
	e.DELETE("/request/:id", taskc.DeleteTask)
	e.GET("/requests", taskc.GetTasks)

	go taskc.Worker.HandlingChan(taskc.ReqChan, taskc.ResChan)

	go func() {
		if err := e.Start(":1323"); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
