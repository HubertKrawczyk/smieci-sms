package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "time"

    "smieci-sms/config"
    "smieci-sms/internal/api"
    "smieci-sms/internal/repository"
    "smieci-sms/internal/scheduler"
    "smieci-sms/internal/service"
)

func main() {
    cfg := config.LoadConfig()

    db, err := repository.NewDB(cfg.DatabaseURL)
    if err != nil {
        log.Fatalf("failed to connect to database: %v", err)
    }
    defer db.Close()

    userRepo := repository.NewUserRepository(db)
    smsService := service.NewSMSService(cfg.SMSProviderAPIKey)
    garbageService := service.NewGarbageService(cfg.CityGarbageURL)
    appScheduler := scheduler.NewScheduler(userRepo, garbageService, smsService)

    router := api.NewRouter(userRepo)
    appScheduler.ScheduleDailyTasks()

    srv := &http.Server{
        Addr:    cfg.ServerAddress,
        Handler: router,
    }

    go func() {
        log.Printf("starting server on %s", cfg.ServerAddress)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server failed: %v", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt)
    <-quit

    log.Println("shutting down server")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("server shutdown failed: %v", err)
    }
}
