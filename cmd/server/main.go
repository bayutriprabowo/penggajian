package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/config"
	"github.com/bayutriprabowo/penggajian/internal/handler"
	"github.com/bayutriprabowo/penggajian/internal/model"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"github.com/bayutriprabowo/penggajian/internal/router"
	"github.com/bayutriprabowo/penggajian/internal/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// main adalah entrypoint aplikasi: memuat konfigurasi, menghubungkan
// database, menjalankan migrasi dan seeder, lalu menjalankan HTTP server
// dengan graceful shutdown.
func main() {
	cfg := config.Load()
	if cfg.JWTSecret == "change-me-in-production" {
		log.Println("[peringatan] JWT_SECRET masih default, ganti untuk produksi")
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		log.Fatalf("gagal koneksi database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("gagal mengambil koneksi sql: %v", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	if err := db.AutoMigrate(
		&model.Role{},
		&model.Permission{},
		&model.User{},
		&model.Employee{},
		&model.Allowance{},
		&model.Overtime{},
		&model.Payroll{},
	); err != nil {
		log.Fatalf("gagal migrasi database: %v", err)
	}

	repos := newRepos(db)
	svc := newServices(repos, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := service.SeedRBAC(ctx, repos.Roles, repos.Users, cfg); err != nil {
		log.Fatalf("gagal seed RBAC: %v", err)
	}
	if cfg.SeedDemo {
		if err := service.SeedDemo(ctx, repos, svc.Auth, svc.Employee, svc.Overtime, svc.Payroll, cfg); err != nil {
			log.Fatalf("gagal seed data demo: %v", err)
		}
	}

	// mode seed saja: selesai setelah migrasi + seeder, tanpa menjalankan server
	if cfg.SeedOnly {
		if err := sqlDB.Close(); err != nil {
			log.Printf("gagal menutup koneksi database: %v", err)
		}
		log.Println("seeding selesai, server tidak dijalankan (SEED_ONLY=true)")
		return
	}

	handlers := &router.Handlers{
		Auth:     handler.NewAuthHandler(svc.Auth),
		User:     handler.NewUserHandler(svc.Auth),
		Employee: handler.NewEmployeeHandler(svc.Employee),
		Overtime: handler.NewOvertimeHandler(svc.Overtime),
		Payroll:  handler.NewPayrollHandler(svc.Payroll),
		Tax:      handler.NewTaxHandler(svc.Tax),
		Health:   handler.NewHealthHandler(sqlDB),
	}

	httpHandler := router.New(handlers, svc.Auth, repos.Users, repos.Roles)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           httpHandler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("server penggajian berjalan di http://localhost:%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server berhenti: %v", err)
		}
	}()

	stopCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-stopCtx.Done()

	log.Println("mematikan server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("gagal shutdown bersih: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("gagal menutup koneksi database: %v", err)
	}
	log.Println("server berhenti")
}

// newRepos membuat seluruh implementasi repository dengan satu koneksi DB.
func newRepos(db *gorm.DB) *repository.Repos {
	return &repository.Repos{
		Users:     repository.NewUserRepository(db),
		Roles:     repository.NewRoleRepository(db),
		Employees: repository.NewEmployeeRepository(db),
		Overtimes: repository.NewOvertimeRepository(db),
		Payrolls:  repository.NewPayrollRepository(db),
	}
}

// newServices menyusun seluruh service dengan dependensi repository dan konfigurasi.
func newServices(repos *repository.Repos, cfg *config.Config) *service.Services {
	auth := service.NewAuthService(repos.Users, repos.Roles, cfg.JWTSecret, cfg.JWTExpireHours)
	return &service.Services{
		Auth:     auth,
		Employee: service.NewEmployeeService(repos.Employees),
		Overtime: service.NewOvertimeService(repos.Overtimes, repos.Employees),
		Payroll:  service.NewPayrollService(repos.Payrolls, repos.Employees, repos.Overtimes, cfg),
		Tax:      service.NewTaxService(repos.Payrolls, repos.Employees),
	}
}
