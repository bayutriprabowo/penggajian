package router

import (
	"log"
	"net/http"
	"time"

	"github.com/bayutriprabowo/penggajian/internal/handler"
	"github.com/bayutriprabowo/penggajian/internal/middleware"
	"github.com/bayutriprabowo/penggajian/internal/repository"
	"github.com/bayutriprabowo/penggajian/internal/service"
)

type Handlers struct {
	Auth     *handler.AuthHandler
	User     *handler.UserHandler
	Employee *handler.EmployeeHandler
	Overtime *handler.OvertimeHandler
	Payroll  *handler.PayrollHandler
	Tax      *handler.TaxHandler
	Health   *handler.HealthHandler
}

// New menyusun seluruh route API dengan middleware auth/RBAC.
func New(
	h *Handlers,
	auth service.AuthService,
	users repository.UserRepository,
	roles repository.RoleRepository,
) http.Handler {
	mux := http.NewServeMux()
	api := http.NewServeMux()

	// publik
	optionalAuth := middleware.OptionalAuth(auth, users, roles)
	api.Handle("POST /auth/register", optionalAuth(http.HandlerFunc(h.Auth.Register)))
	api.HandleFunc("POST /auth/login", h.Auth.Login)

	// butuh autentikasi
	authMW := middleware.Auth(auth, users, roles)

	api.Handle("GET /users/me", authMW(middleware.RequirePermission("user:read")(http.HandlerFunc(h.User.Me))))
	api.Handle("GET /users", authMW(middleware.RequirePermission("user:list")(http.HandlerFunc(h.User.List))))

	api.Handle("POST /employees", authMW(middleware.RequirePermission("employee:write")(http.HandlerFunc(h.Employee.Create))))
	api.Handle("GET /employees", authMW(middleware.RequirePermission("employee:read")(http.HandlerFunc(h.Employee.List))))
	api.Handle("GET /employees/{id}", authMW(middleware.RequirePermission("employee:read")(http.HandlerFunc(h.Employee.GetByID))))
	api.Handle("PUT /employees/{id}", authMW(middleware.RequirePermission("employee:write")(http.HandlerFunc(h.Employee.Update))))
	api.Handle("DELETE /employees/{id}", authMW(middleware.RequirePermission("employee:write")(http.HandlerFunc(h.Employee.Delete))))

	api.Handle("POST /overtimes", authMW(middleware.RequirePermission("overtime:write")(http.HandlerFunc(h.Overtime.Create))))
	api.Handle("GET /overtimes/employee/{id}", authMW(middleware.RequirePermission("employee:read")(http.HandlerFunc(h.Overtime.ListByEmployee))))

	api.Handle("POST /payrolls/run", authMW(middleware.RequirePermission("payroll:run")(http.HandlerFunc(h.Payroll.Run))))
	api.Handle("GET /payrolls", authMW(middleware.RequirePermission("payroll:read")(http.HandlerFunc(h.Payroll.ListByPeriod))))
	api.Handle("GET /payrolls/{id}", authMW(middleware.RequirePermission("payroll:read")(http.HandlerFunc(h.Payroll.GetByID))))
	api.Handle("GET /payrolls/employee/{id}", authMW(middleware.RequirePermission("payroll:read")(http.HandlerFunc(h.Payroll.ListByEmployee))))
	api.Handle("PATCH /payrolls/{id}/approve", authMW(middleware.RequirePermission("payroll:approve")(http.HandlerFunc(h.Payroll.Approve))))
	api.Handle("PATCH /payrolls/{id}/paid", authMW(middleware.RequirePermission("payroll:approve")(http.HandlerFunc(h.Payroll.MarkPaid))))

	api.Handle("GET /thr/calculate", authMW(middleware.RequirePermission("payroll:read")(http.HandlerFunc(h.Payroll.CalculateTHR))))

	api.Handle("GET /tax/ter", authMW(middleware.RequirePermission("tax:read")(http.HandlerFunc(h.Tax.TERTables))))
	api.Handle("GET /tax/ter-info", authMW(middleware.RequirePermission("tax:read")(http.HandlerFunc(h.Tax.TERInfo))))
	api.Handle("GET /tax/annual-recap", authMW(middleware.RequirePermission("tax:read")(http.HandlerFunc(h.Tax.AnnualRecap))))

	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))
	mux.HandleFunc("GET /health", h.Health.Check)

	return recoverer(logging(mux))
}

// recoverer menangkap panic handler agar server tidak berhenti.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[panic] %v (%s %s)", rec, r.Method, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"status":"error","message":"terjadi kesalahan server"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// logging mencatat method, path, dan durasi setiap request.
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}
